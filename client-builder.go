package httpclient

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/cocotyty/cookiejar"
	"golang.org/x/net/proxy"
	"golang.org/x/net/publicsuffix"
)

var emptyDuration time.Duration

const defaultCachePrefix = "http/"

const defaultCachedTime = 10 * time.Minute

type StoreCookie func(sessionID string, jar http.CookieJar)

type HttpService Builder

func (s *HttpService) Get(sessionid string) *HttpRequest {
	return (*Builder)(s).Get(sessionid)
}

func (s *HttpService) Post(sessionid string) *HttpRequest {
	return (*Builder)(s).Post(sessionid)
}

type Builder struct {
	SessionCachePrefix string
	SessionCachedTime  time.Duration
	Timeout            time.Duration
	Proxy              string
	Auth               *proxy.Auth
	Cache              Cache // Cache to store cookies
	UserAgentsPool     []string
	Transport          *http.Transport
	once               sync.Once
}

func (builder *Builder) loadCache(sessionID string) (data []byte) {
	prefix := builder.SessionCachePrefix
	if prefix == "" {
		prefix = defaultCachePrefix
	}
	if cacheData, found := builder.Cache.Get(prefix + sessionID); found && cacheData != nil {
		data = cacheData.([]byte)
		return
	}
	return
}

func (builder *Builder) saveCache(sessionID string, data interface{}) {
	prefix := builder.SessionCachePrefix
	if prefix == "" {
		prefix = defaultCachePrefix
	}
	cachedTime := builder.SessionCachedTime
	if cachedTime == emptyDuration {
		cachedTime = defaultCachedTime
	}
	builder.Cache.Set(defaultCachePrefix+sessionID, data, cachedTime)
	return
}

func (builder *Builder) initTransport() {
	if builder.Transport == nil {
		builder.Transport = &http.Transport{
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			MaxConnsPerHost:       2000,
			MaxIdleConns:          2000,
			IdleConnTimeout:       time.Minute,
			ExpectContinueTimeout: 0,
		}
	}
	dialer := &net.Dialer{Timeout: builder.Timeout}
	if builder.Proxy == "" {
		builder.Transport.DialContext = dialer.DialContext
	} else {
		proxyDialer, _ := proxy.SOCKS5("tcp", builder.Proxy, builder.Auth, dialer)
		builder.Transport.DialContext = proxyDialer.(proxy.ContextDialer).DialContext
	}
}

func (builder *Builder) newRequest(sessionID string, noAutoRedirect bool) *HttpRequest {
	builder.once.Do(builder.initTransport)
	jarData := builder.loadCache(sessionID)
	return NewHttpRequest(builder.injectCookiesClient(jarData, noAutoRedirect)).
		SetCookieStore(builder.storeCookie).
		SetUserAgentPool(builder.UserAgentsPool).
		Session(sessionID)
}

func (builder *Builder) NoAutoRedirectRequest(sessionID string) *HttpRequest {
	return builder.newRequest(sessionID, true)
}

func (builder *Builder) Request(sessionID string) *HttpRequest {
	return builder.newRequest(sessionID, false)
}

func (builder *Builder) Get(sessionID string) *HttpRequest {
	return builder.Request(sessionID).Get()
}
func (builder *Builder) Post(sessionID string) *HttpRequest {
	return builder.Request(sessionID).Post()
}

func (builder *Builder) storeCookie(sessionID string, cookieJar http.CookieJar) {
	data, _ := json.Marshal(cookieJar)
	builder.saveCache(sessionID, data)
}

func (builder *Builder) injectCookiesClient(cookieJarBytes []byte, noAutoRedirect bool) *http.Client {
	var cl = &http.Client{}
	cl.Transport = builder.Transport
	cl.Timeout = builder.Timeout
	cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if noAutoRedirect {
			return http.ErrUseLastResponse
		}
		return nil
	}
	if cookieJarBytes == nil {
		Jars, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
		}
		cl.Jar = Jars
	} else {
		Jars, err := cookiejar.LoadFromJson(&cookiejar.Options{PublicSuffixList: publicsuffix.List}, cookieJarBytes)
		if err != nil {
		}
		cl.Jar = Jars
	}

	return cl
}
