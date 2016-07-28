package httpclient

import "net/http"

func Get(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, method:http.MethodGet, client:http.DefaultClient}
}
func Post(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, method:http.MethodPost, client:http.DefaultClient}
}
func Delete(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, method:http.MethodDelete, client:http.DefaultClient}
}
func Put(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, method:http.MethodPut, client:http.DefaultClient}
}
func Patch(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, method:http.MethodPatch, client:http.DefaultClient}
}
func Head(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, method:http.MethodHead, client:http.DefaultClient}
}
func Options(url string) *HttpRequest {
	return &HttpRequest{header:http.Header{}, method:http.MethodOptions, client:http.DefaultClient}
}