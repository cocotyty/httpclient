package httpclient

import (
	"crypto/tls"
	"github.com/cocotyty/cookiejar"
	"github.com/cocotyty/json"
	logger "github.com/golang/glog"
	"golang.org/x/net/proxy"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
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
		cachedTime = builder.SessionCachedTime
	}
	builder.Cache.Set(defaultCachePrefix+sessionID, data, cachedTime)
	return
}

func (builder *Builder) Request(sessionID string) *HttpRequest {
	jarData := builder.loadCache(sessionID)
	return NewHttpRequest(builder.makeCookieClient(jarData)).
		SetCookieStore(builder.storeCookie).
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
	var cl *http.Client
	dialer := &net.Dialer{Timeout: builder.Timeout}
	if builder.Proxy == "" {
		cl = &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial:            dialer.Dial,
		}}
	} else {
		proxyDialer, _ := proxy.SOCKS5("tcp", builder.Proxy, builder.Auth, dialer)
		cl = &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial:            proxyDialer.Dial,
		}}
	}
	cl.Timeout = builder.Timeout
	if tr, ok := cl.Transport.(*http.Transport); ok {
		tr.ExpectContinueTimeout = 0
	}
	cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		logger.Info(req.URL)
		return nil
	}
	if cookieJarBytes == nil {
		Jars, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			logger.Info("[cookie-jar-err]", err)
		}
		cl.Jar = Jars
	} else {
		Jars, err := cookiejar.LoadFromJson(&cookiejar.Options{PublicSuffixList: publicsuffix.List}, cookieJarBytes)
		if err != nil {
			logger.Info("[cookie-jar-err]", err)
		}
		cl.Jar = Jars
	}

	return cl
}
