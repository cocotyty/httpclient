package httpclient

import (
	"os"
	"net/http"
	"fmt"
	"encoding/json"
	"time"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"crypto/tls"
	"net"
	"golang.org/x/net/publicsuffix"
	"github.com/patrickmn/go-cache"
	"log"
	"bytes"
	"net/url"
	"github.com/cocotyty/cookiejar"
)

var logger = log.New(os.Stderr, "[http]", log.Ldate | log.Ltime | log.Llongfile)

type HttpService struct {
	Cache *cache.Cache
}

func (this *HttpService)Get(sessionid string) *HttpRequest {
	var jar []byte
	if data, found := this.Cache.Get("http/" + sessionid); found&&data != nil {
		jar = data.([]byte)
	}
	return &HttpRequest{header:http.Header{}, method:"GET", sessionID:sessionid, service:this, client:clientWithCookieJson(jar)}
}
func (this *HttpService)Post(sessionid string) *HttpRequest {
	var jar []byte
	if data, found := this.Cache.Get("http/" + sessionid); found&&data != nil {
		jar = data.([]byte)
	}
	fmt.Println(string(jar))
	return &HttpRequest{header:http.Header{}, method:"POST", sessionID:sessionid, service:this, client:clientWithCookieJson(jar)}
}

func (this *HttpService)saveCookie(sessionID string, cookieJar interface{}) {
	data, _ := json.Marshal(cookieJar)
	this.Cache.Set("http/" + sessionID, data, time.Minute * 60)
}

type HttpRequest struct {
	method    string
	url       string
	gb18030   bool
	header    http.Header
	body      []byte
	jsonData  interface{}
	querys    map[string][]string
	params    map[string][]string
	client    *http.Client
	service   *HttpService
	sessionID string
}
type HttpResponse struct {
	header http.Header
	body   []byte
}

func (resp *HttpResponse)Body() []byte {
	return resp.body
}

func (resp *HttpResponse)String() string {
	return string(resp.body)
}
func (resp *HttpResponse)GB18030() string {
	data, err := simplifiedchinese.GB18030.NewDecoder().Bytes(resp.body)
	if err != nil {
		return string(resp.body)
	}
	return string(data)
}
func (req *HttpRequest)Url(url string) *HttpRequest {
	req.url = url
	return req
}
func (req *HttpRequest)Query(k, v string) *HttpRequest {
	if req.querys == nil {
		req.querys = make(map[string][]string)
	}
	req.querys[k] = []string{v}
	return req
}
func (req *HttpRequest)QueryArray(k string, v []string) *HttpRequest {
	if req.querys == nil {
		req.querys = make(map[string][]string)
	}
	req.querys[k] = v
	return req
}
func (req *HttpRequest)Param(k string, v string) *HttpRequest {
	if req.params == nil {
		req.params = make(map[string][]string)
		req.Head("Content-Type", "application/x-www-form-urlencoded")
	}
	req.params[k] = []string{v}
	return req
}
func (req *HttpRequest)ParamArray(k string, v []string) *HttpRequest {
	if req.params == nil {
		req.params = make(map[string][]string)
		req.Head("Content-Type", "application/x-www-form-urlencoded")
	}
	req.params[k] = v
	return req
}
func (req *HttpRequest)JSON(data interface{}) *HttpRequest {
	req.jsonData = data
	return req
}
func (req *HttpRequest)Body(body []byte) *HttpRequest {
	req.body = body
	return req
}

func (req *HttpRequest)Head(k, v string) *HttpRequest {
	req.header.Set(k, v)
	return req
}
func (req *HttpRequest)GB18030() *HttpRequest {
	req.gb18030 = true
	return req
}

func (req *HttpRequest)UTF8() *HttpRequest {
	req.gb18030 = false
	return req
}
func (req *HttpRequest)Send() (resp *HttpResponse, err error) {
	if req.querys != nil {
		req.url = req.url + "?" + string(buildEncoded(req.querys, req.gb18030))
	}
	if req.params != nil {
		req.body = buildEncoded(req.params, req.gb18030)
	}
	if req.jsonData != nil {
		req.body, err = json.Marshal(req.jsonData)
		if err != nil {
			return nil, err
		}
	}
	logger.Println(req.header, string(req.body), req.url, req.querys, req.method)
	hrep, err := http.NewRequest(req.method, req.url, bytes.NewReader(req.body))
	if err != nil {
		return nil, err
	}
	hrep.Header = req.header
	hresp, err := req.client.Do(hrep)
	if hresp != nil&& hresp.Body != nil {
		defer hresp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(hresp.Body)
	if err != nil {
		return nil, err
	}
	req.service.saveCookie(req.sessionID, req.client.Jar)
	return &HttpResponse{body:data, header:hresp.Header}, nil
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
	if result[len(result) - 1] == '&' {
		result = result[:len(result) - 1]
	}
	return result
}
func clientWithCookieJson(src []byte) *http.Client {
	cl := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify:true},
		Dial: func(netw, addr string) (net.Conn, error) {
			d := &net.Dialer{Timeout: time.Second * 90}
			return d.Dial(netw, addr)
		},
	}}
	if src == nil {
		Jars, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			logger.Println("[cookie-jar-err]", err)
		}
		cl.Jar = Jars
	} else {
		Jars, err := cookiejar.LoadFromJson(&cookiejar.Options{PublicSuffixList: publicsuffix.List}, src)
		if err != nil {
			logger.Println("[cookie-jar-err]", err)
		}
		cl.Jar = Jars
	}

	return cl
}

