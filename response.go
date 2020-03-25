package httpclient

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
)

type HttpResponse struct {
	code     int
	err      error
	header   http.Header
	body     []byte
	url      *url.URL
	encoding encoding.Encoding
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
	if resp.encoding == nil {
		return string(resp.body), nil
	}
	data, err := resp.encoding.NewDecoder().Bytes(resp.body)
	if err != nil {
		return "", err
	}
	return string(data), err
}

func (resp *HttpResponse) JSON(data interface{}) error {
	if resp.err != nil {
		return resp.err
	}
	return json.Unmarshal(resp.body, data)
}

// http://www.w3.org/TR/encoding
func (resp *HttpResponse) Encoding(name string) *HttpResponse {
	if resp.err != nil {
		return resp
	}
	enc, err := htmlindex.Get(name)
	if err != nil {
		resp.err = err
		return resp
	}
	resp.encoding = enc
	return resp
}
