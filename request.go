package httpclient

import (
	"bytes"
	"encoding/json"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/cocotyty/cookiejar"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
)

// from my mac
const defaultUA = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36`

type HttpRequest struct {
	err            error
	method         string
	url            string
	host           string
	encoding       encoding.Encoding
	header         http.Header
	body           []byte
	jsonData       interface{}
	querys         [][]string
	params         map[string][]string
	client         *http.Client
	storeCookie    StoreCookie
	sessionID      string
	cookies        []*http.Cookie
	UserAgentsPool []string
	dumpRequest    io.WriteCloser
	dumpResponse   io.WriteCloser
}

func NewHttpRequest(client *http.Client) *HttpRequest {
	return &HttpRequest{header: http.Header{}, client: client}
}
func (req *HttpRequest) SetCookieStore(store StoreCookie) *HttpRequest {
	req.storeCookie = store
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

func (req *HttpRequest) Put() *HttpRequest {
	req.method = http.MethodPut
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

func (req *HttpRequest) Cookies(raw string) *HttpRequest {
	var request http.Request
	request.Header = http.Header{"Cookie": []string{raw}}
	cookies := request.Cookies()
	req.cookies = append(req.cookies, cookies...)
	return req
}

func (req *HttpRequest) AddCookie(ck *http.Cookie) *HttpRequest {
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
	_, _ = hash.Write([]byte(req.sessionID))
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

func (req *HttpRequest) Encoding(name string) *HttpRequest {
	req.encoding, req.err = htmlindex.Get(name)
	return req
}

func (req *HttpRequest) Dump() *HttpRequest {
	req.dumpRequest = &dumpPrinter{firstLine: ">> request >>"}
	req.dumpResponse = &dumpPrinter{firstLine: "<< response <<"}
	return req
}

type dumpPrinter struct {
	firstLine string
	bytes.Buffer
}

func (d *dumpPrinter) Close() error {
	lines := bytes.Split(d.Buffer.Bytes(), []byte("\n"))
	os.Stderr.WriteString(d.firstLine)
	os.Stderr.Write([]byte{'\n'})
	for _, line := range lines {
		os.Stderr.Write(line)
		os.Stderr.Write([]byte{'\n'})
	}
	return nil
}

func (req *HttpRequest) DumpTo(request io.WriteCloser, response io.WriteCloser) *HttpRequest {
	req.dumpRequest = request
	req.dumpResponse = response
	return req
}

func (req *HttpRequest) Send() (resp *HttpResponse) {
	if req.err != nil {
		return &HttpResponse{err: req.err}
	}
	resp = &HttpResponse{}
	var err error
	if req.querys != nil {
		req.url = req.url + "?" + string(encodeQuery(req.querys, req.encoding))
	}
	if req.params != nil {
		req.body = encodeForm(req.params, req.encoding)
	}
	if req.jsonData != nil {
		req.header.Set("Content-Type", "application/json")
		req.body, err = json.Marshal(req.jsonData)
		if err != nil {
			resp.err = err
			return
		}
	}
	request, err := http.NewRequest(req.method, req.url, bytes.NewReader(req.body))
	if err != nil {
		resp.err = err
		return
	}
	if req.host != "" {
		request.Host = req.host
	}
	if req.cookies != nil {
		if req.client.Jar == nil {
			req.client.Jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		}
		req.client.Jar.SetCookies(request.URL, req.cookies)
	}
	request.Header = req.header

	if req.dumpRequest != nil {
		data, _ := httputil.DumpRequest(request, true)
		_, _ = req.dumpRequest.Write(data)
		_ = req.dumpRequest.Close()
	}

	response, err := req.client.Do(request)

	if req.dumpResponse != nil && response != nil {
		data, _ := httputil.DumpResponse(response, true)
		_, _ = req.dumpResponse.Write(data)
		_ = req.dumpResponse.Close()
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	if err != nil {
		resp.err = err
		return
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		resp.err = err
		return
	}
	if req.storeCookie != nil {
		req.storeCookie(req.sessionID, req.client.Jar)
	}
	return &HttpResponse{code: response.StatusCode, body: data, header: response.Header, url: response.Request.URL}
}
