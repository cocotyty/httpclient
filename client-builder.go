package httpclient

import (
	"crypto/tls"
	"github.com/cocotyty/cookiejar"
	"github.com/cocotyty/json"
	logger "github.com/cocotyty/mlog"
	"golang.org/x/net/proxy"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
	"sync"
	"time"
)

var emptyDuration time.Duration

const defaultCachePrefix = "http/"

const defaultCachedTime = 10 * time.Minute

type StoreCookie func(sessionID string, jar http.CookieJar)

type HttpService Builder

func (hs *HttpService) Get(sessionid string) *HttpRequest {
	return (*Builder)(hs).Get(sessionid)
}

func (hs *HttpService) Post(sessionid string) *HttpRequest {
	return (*Builder)(hs).Post(sessionid)
}

type Builder struct {
	SessionCachePrefix string
	SessionCachedTime  time.Duration
	Timeout            time.Duration
	Proxy              string
	Auth               *proxy.Auth
	Cache              Cache
	UserAgentsPool     []string
	transport          *http.Transport
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
	builder.transport = &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:       2000,
		MaxIdleConns:          2000,
		IdleConnTimeout:       time.Minute,
		ExpectContinueTimeout: 0,
	}
	dialer := &net.Dialer{Timeout: builder.Timeout}
	if builder.Proxy == "" {
		builder.transport.DialContext = dialer.DialContext
	} else {
		proxyDialer, _ := proxy.SOCKS5("tcp", builder.Proxy, builder.Auth, dialer)
		builder.transport.DialContext = proxyDialer.(proxy.ContextDialer).DialContext
	}
}

func (builder *Builder) Request(sessionID string) *HttpRequest {
	builder.once.Do(builder.initTransport)
	jarData := builder.loadCache(sessionID)
	return NewHttpRequest(builder.makeCookieClient(jarData)).
		SetCookieStore(builder.storeCookie).
		SetUserAgentPool(builder.UserAgentsPool).
		Session(sessionID)
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

func (builder *Builder) makeCookieClient(cookieJarBytes []byte) *http.Client {
	var cl = &http.Client{}
	cl.Transport = builder.transport
	cl.Timeout = builder.Timeout
	cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		logger.Debug(req.URL)
		return nil
	}
	if cookieJarBytes == nil {
		Jars, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			logger.Debug("[cookie-jar-err]", err)
		}
		cl.Jar = Jars
	} else {
		Jars, err := cookiejar.LoadFromJson(&cookiejar.Options{PublicSuffixList: publicsuffix.List}, cookieJarBytes)
		if err != nil {
			logger.Debug("[cookie-jar-err]", err)
		}
		cl.Jar = Jars
	}

	return cl
}
