package httpclient

import (
	logger "github.com/cocotyty/mlog"

	"bytes"
	"github.com/cocotyty/json"
	"hash/fnv"
	"io/ioutil"
	"net/http"
)

// from my mac
const defaultUA = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36`

type HttpRequest struct {
	method         string
	url            string
	host           string
	gb18030        bool
	header         http.Header
	body           []byte
	jsonData       interface{}
	querys         [][]string
	params         map[string][]string
	client         *http.Client
	storeCookieFn  StoreCookie
	sessionID      string
	cookies        []*http.Cookie
	UserAgentsPool []string
}

func NewHttpRequest(client *http.Client) *HttpRequest {
	return &HttpRequest{header: http.Header{}, client: client}
}
func (req *HttpRequest) SetCookieStore(store StoreCookie) *HttpRequest {
	req.storeCookieFn = store
	return req
}
func (req *HttpRequest) SetUserAgentPool(uas []string) *HttpRequest {
	req.UserAgentsPool = uas
	return req
}
func (req *HttpRequest) Session(sessionID string) *HttpRequest {
	req.sessionID = sessionID
	return req
}
func (req *HttpRequest) Method(method string) *HttpRequest {
	req.method = method
	return req
}

func (req *HttpRequest) Get() *HttpRequest {
	req.method = http.MethodGet
	return req
}

func (req *HttpRequest) Post() *HttpRequest {
	req.method = http.MethodPost
	return req
}

func (req *HttpRequest) Patch() *HttpRequest {
	req.method = http.MethodPatch
	return req
}

func (req *HttpRequest) Connect() *HttpRequest {
	req.method = http.MethodConnect
	return req
}

func (req *HttpRequest) Delete() *HttpRequest {
	req.method = http.MethodDelete
	return req
}
func (req *HttpRequest) MethodHead() *HttpRequest {
	req.method = http.MethodHead
	return req
}
func (req *HttpRequest) Options() *HttpRequest {
	req.method = http.MethodOptions
	return req
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
func (req *HttpRequest) RefererInHeader(refer string) *HttpRequest {
	return req.Head("Referer", refer)
}

func (req *HttpRequest) Host(host string) *HttpRequest {
	req.host = host
	return req
}

func (req *HttpRequest) UserAgentInHeader(userAgent string) *HttpRequest {
	return req.Head("User-Agent", userAgent)
}

func (req *HttpRequest) AutoSelectUserAgent() *HttpRequest {
	if req.UserAgentsPool == nil || len(req.UserAgentsPool) == 0 {
		return req.UserAgentInHeader(defaultUA)
	}

	hash := fnv.New64()
	hash.Write([]byte(req.sessionID))
	index := hash.Sum64()

	return req.UserAgentInHeader(req.UserAgentsPool[index%uint64(len(req.UserAgentsPool))])
}
func (req *HttpRequest) OriginInHeader(origin string) *HttpRequest {
	return req.Head("Origin", origin)
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
	if req.host != "" {
		hrep.Host = req.host
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
	if req.storeCookieFn != nil {
		req.storeCookieFn(req.sessionID, req.client.Jar)
	}
	return &HttpResponse{code: hresp.StatusCode, body: data, header: hresp.Header, url: hresp.Request.URL}
}
