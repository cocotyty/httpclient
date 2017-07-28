package httpclient

import (
	"crypto/tls"
	"net"
	"net/http"
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
func New(cl *http.Client) *client {
	return &client{cl: cl}
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
