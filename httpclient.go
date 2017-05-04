package httpclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/cocotyty/cookiejar"
	"github.com/cocotyty/summer"
	"golang.org/x/net/proxy"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

type client struct {
	cl *http.Client
}

func NewNoSSLVerify() *client {
	return &client{cl: &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	},
	}
}
func New(client *http.Client) *client {
	return &client{cl: client}
}

func (cl *client) Get(url string) *HttpRequest {
	return &HttpRequest{header: http.Header{}, url: url, method: http.MethodGet, client: cl.cl}
}
func (cl *client) Post(url string) *HttpRequest {
	return &HttpRequest{header: http.Header{}, url: url, method: http.MethodPost, client: cl.cl}
}
func (cl *client) Delete(url string) *HttpRequest {
	return &HttpRequest{header: http.Header{}, url: url, method: http.MethodDelete, client: cl.cl}
}
func (cl *client) Put(url string) *HttpRequest {
	return &HttpRequest{header: http.Header{}, url: url, method: http.MethodPut, client: cl.cl}
}
func (cl *client) Patch(url string) *HttpRequest {
	return &HttpRequest{header: http.Header{}, url: url, method: http.MethodPatch, client: cl.cl}
}
func (cl *client) Head(url string) *HttpRequest {
	return &HttpRequest{header: http.Header{}, url: url, method: http.MethodHead, client: cl.cl}
}
func (cl *client) Options(url string) *HttpRequest {
	return &HttpRequest{header: http.Header{}, url: url, method: http.MethodOptions, client: cl.cl}
}

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, exp time.Duration)
}

var logger = summer.NewSimpleLog("http", summer.DebugLevel)

type HttpService struct {
	Timeout time.Duration
	Proxy   string
	Auth    *proxy.Auth
	Cache   Cache
}

func (this *HttpService) Get(sessionid string) *HttpRequest {
	var jar []byte
	if data, found := this.Cache.Get("http/" + sessionid); found && data != nil {
		jar = data.([]byte)
	}
	logger.Debug("[cookies]load from cache", string(jar))
	return &HttpRequest{header: http.Header{}, method: "GET", sessionID: sessionid, service: this, client: this.clientWithCookieJson(jar)}
}
func (this *HttpService) Post(sessionid string) *HttpRequest {
	var jar []byte
	if data, found := this.Cache.Get("http/" + sessionid); found && data != nil {
		jar = data.([]byte)
	}
	logger.Debug("[cookies]load from cache", string(jar))
	return &HttpRequest{header: http.Header{}, method: "POST", sessionID: sessionid, service: this, client: this.clientWithCookieJson(jar)}
}

func (this *HttpService) saveCookie(sessionID string, cookieJar interface{}) {
	data, _ := json.Marshal(cookieJar)
	logger.Debug("[cookies]save to cache", string(data))
	this.Cache.Set("http/"+sessionID, data, time.Minute*60)
}

type HttpRequest struct {
	method    string
	url       string
	gb18030   bool
	header    http.Header
	body      []byte
	jsonData  interface{}
	querys    [][]string
	params    map[string][]string
	client    *http.Client
	service   *HttpService
	sessionID string
	cookies   []*http.Cookie
}
type HttpResponse struct {
	code   int
	err    error
	header http.Header
	body   []byte
	url    *url.URL
}

func (resp *HttpResponse) Code() (int, error) {
	if resp.err != nil {
		return 0, resp.err
	}
	return resp.code, nil
}
func (resp *HttpResponse) Body() ([]byte, error) {
	if resp.err != nil {
		return nil, resp.err
	}
	return resp.body, nil
}
func (resp *HttpResponse) Header() http.Header {
	return resp.header
}
func (resp *HttpResponse) URL() *url.URL {
	return resp.url
}
func (resp *HttpResponse) String() (string, error) {
	if resp.err != nil {
		return "", resp.err
	}
	return string(resp.body), nil
}
func (resp *HttpResponse) JSON(data interface{}) error {
	if resp.err != nil {
		return resp.err
	}
	return json.Unmarshal(resp.body, data)
}
func (resp *HttpResponse) GB18030() (string, error) {
	if resp.err != nil {
		return "", resp.err
	}
	data, err := simplifiedchinese.GB18030.NewDecoder().Bytes(resp.body)
	if err != nil {
		return string(resp.body), nil
	}
	return string(data), nil
}
func (req *HttpRequest) AddCookie(ck *http.Cookie) *HttpRequest {
	if req.cookies == nil {
		req.cookies = []*http.Cookie{ck}
		return req
	}
	req.cookies = append(req.cookies, ck)
	return req
}
func (req *HttpRequest) Url(url string) *HttpRequest {
	req.url = url
	return req
}
func (req *HttpRequest) Query(k, v string) *HttpRequest {
	if req.querys == nil {
		req.querys = [][]string{}
	}
	req.querys = append(req.querys, []string{k, v})
	return req
}
func (req *HttpRequest) QueryArray(k string, vs []string) *HttpRequest {
	if req.querys == nil {
		req.querys = [][]string{}
	}
	for _, v := range vs {
		req.querys = append(req.querys, []string{k, v})
	}
	return req
}
func (req *HttpRequest) Param(k string, v string) *HttpRequest {
	if req.params == nil {
		req.params = make(map[string][]string)
		req.Head("Content-Type", "application/x-www-form-urlencoded")
	}
	req.params[k] = []string{v}
	return req
}
func (req *HttpRequest) ParamArray(k string, v []string) *HttpRequest {
	if req.params == nil {
		req.params = make(map[string][]string)
		req.Head("Content-Type", "application/x-www-form-urlencoded")
	}
	req.params[k] = v
	return req
}
func (req *HttpRequest) JSON(data interface{}) *HttpRequest {
	req.jsonData = data
	return req
}
func (req *HttpRequest) Body(body []byte) *HttpRequest {
	req.body = body
	return req
}

func (req *HttpRequest) Head(k, v string) *HttpRequest {
	req.header.Set(k, v)
	return req
}

func (req *HttpRequest) GB18030() *HttpRequest {
	req.gb18030 = true
	return req
}

func (req *HttpRequest) UTF8() *HttpRequest {
	req.gb18030 = false
	return req
}
func (req *HttpRequest) Send() (resp *HttpResponse) {
	resp = &HttpResponse{}
	var err error
	if req.querys != nil {
		req.url = req.url + "?" + string(buildQueryEncoded(req.querys, req.gb18030))
	}
	logger.Debug(req.url)
	if req.params != nil {
		req.body = buildEncoded(req.params, req.gb18030)
	}
	if req.jsonData != nil {
		req.body, err = json.Marshal(req.jsonData)
		if err != nil {
			resp.err = err
			return
		}
	}
	hrep, err := http.NewRequest(req.method, req.url, bytes.NewReader(req.body))
	if err != nil {
		resp.err = err
		return
	}
	if req.cookies != nil {
		req.client.Jar.SetCookies(hrep.URL, req.cookies)
	}
	hrep.Header = req.header
	hresp, err := req.client.Do(hrep)
	if hresp != nil && hresp.Body != nil {
		defer hresp.Body.Close()
	}
	if err != nil {
		resp.err = err
		return
	}

	data, err := ioutil.ReadAll(hresp.Body)
	if err != nil {
		resp.err = err
		return
	}
	if req.service != nil {
		req.service.saveCookie(req.sessionID, req.client.Jar)
	}
	return &HttpResponse{code: hresp.StatusCode, body: data, header: hresp.Header, url: hresp.Request.URL}
}

func buildQueryEncoded(source [][]string, gb18030 bool) []byte {
	buf := bytes.NewBuffer(nil)
	for _, kv := range source {
		k, v := kv[0], kv[1]
		buf.WriteString(k)
		buf.WriteByte('=')
		if gb18030 {
			v, _ = simplifiedchinese.GB18030.NewEncoder().String(v)
		}
		buf.WriteString(url.QueryEscape(v))
		buf.WriteByte('&')
	}
	result := buf.Bytes()
	if result[len(result)-1] == '&' {
		result = result[:len(result)-1]
	}
	return result
}
func buildEncoded(source map[string][]string, gb18030 bool) []byte {
	buf := bytes.NewBuffer(nil)
	for k, strs := range source {
		for _, v := range strs {
			buf.WriteString(k)
			buf.WriteByte('=')
			if gb18030 {
				v, _ = simplifiedchinese.GB18030.NewEncoder().String(v)
			}
			buf.WriteString(url.QueryEscape(v))
			buf.WriteByte('&')
		}
	}
	result := buf.Bytes()
	if result[len(result)-1] == '&' {
		result = result[:len(result)-1]
	}
	return result
}

func (this *HttpService) clientWithCookieJson(src []byte) *http.Client {
	var cl *http.Client
	dialer := &net.Dialer{Timeout: time.Second * 90}

	if this.Proxy == "" {
		cl = &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial:            dialer.Dial,
		}}
	} else {
		proxyDialer, _ := proxy.SOCKS5("tcp", this.Proxy, this.Auth, dialer)
		cl = &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial:            proxyDialer.Dial,
		}}
	}
	cl.Timeout = this.Timeout
	if tr, ok := cl.Transport.(*http.Transport); ok {
		tr.ExpectContinueTimeout = 0
	}
	cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		logger.Debug(req.URL)
		return nil
	}
	if src == nil {
		Jars, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			logger.Debug("[cookie-jar-err]", err)
		}
		cl.Jar = Jars
	} else {
		Jars, err := cookiejar.LoadFromJson(&cookiejar.Options{PublicSuffixList: publicsuffix.List}, src)
		if err != nil {
			logger.Debug("[cookie-jar-err]", err)
		}
		cl.Jar = Jars
	}

	return cl
}
