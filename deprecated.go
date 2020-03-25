package httpclient

import (
	"golang.org/x/text/encoding/simplifiedchinese"
)

// Deprecated: Should use `Encoding` method
func (req *HttpRequest) GB18030() *HttpRequest {
	req.encoding = simplifiedchinese.GB18030
	return req
}
// Deprecated: Should use `Encoding` method
func (req *HttpRequest) UTF8() *HttpRequest {
	req.encoding = nil
	return req
}

// Deprecated: Should use `Encoding` method
func (resp *HttpResponse) GB18030() (string, error) {
	if resp.err != nil {
		return "", resp.err
	}
	resp.encoding = simplifiedchinese.GB18030
	return resp.String()
}

// Deprecated: Should use `Encoding` method
func (resp *HttpResponse) Encode(name string) *HttpResponse {
	return resp.Encoding(name)
}
