package httpclient

import "net/http"

func Get(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, url:url, method:http.MethodGet, client:http.DefaultClient}
}
func Post(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, url:url, method:http.MethodPost, client:http.DefaultClient}
}
func Delete(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, url:url, method:http.MethodDelete, client:http.DefaultClient}
}
func Put(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, url:url, method:http.MethodPut, client:http.DefaultClient}
}
func Patch(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, url:url, method:http.MethodPatch, client:http.DefaultClient}
}
func Head(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, url:url, method:http.MethodHead, client:http.DefaultClient}
}
func Options(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, url:url, method:http.MethodOptions, client:http.DefaultClient}
}