package httpclient

import (
	"bytes"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/simplifiedchinese"
	"net/http"
	"net/url"
	"regexp"
)

var contentTypeMatchCharset = regexp.MustCompile(`\bcharset=([\w|\-]*)`)

type HttpResponse struct {
	code     int
	err      error
	header   http.Header
	body     []byte
	url      *url.URL
	encoding encoding.Encoding
}

func (resp *HttpResponse) HTML() (doc *goquery.Document, err error) {
	if resp.err != nil {
		return nil, resp.err
	}
	enc := resp.encoding
	data := resp.body
	if enc != nil {
		data, err = enc.NewDecoder().Bytes(resp.body)
		if err != nil {
			return nil, err
		}
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(data))
}

func (resp *HttpResponse) HTMLDetectedEncode() (doc *goquery.Document, err error) {
	if resp.err != nil {
		return nil, resp.err
	}
	enc := resp.encoding
	if enc == nil {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.body))
		if err != nil {
			return nil, err
		}
		encName := ""
		selector := doc.Find("meta[charset]")
		if selector != nil && selector.Size() > 0 {
			encName, _ = selector.Attr("charset")
			glog.Info("find html encode ", encName)
		} else {
			selector = doc.Find(`meta[http-equiv="Content-Type"]`)
			if selector != nil && selector.Size() > 0 {
				attrContent, exists := selector.Attr("content")
				glog.Info("find html encode from content ", attrContent)
				if exists {
					subMatch := contentTypeMatchCharset.FindStringSubmatch(attrContent)
					if len(subMatch) == 2 {
						glog.Info("find html encode ", subMatch[1])
						encName = subMatch[1]
					}
				}
			}
		}
		if encName != "" {
			enc, err = htmlindex.Get(encName)
			if err != nil {
				return nil, err
			}
		}
	}

	data := resp.body

	if enc != nil {
		data, err = enc.NewDecoder().Bytes(resp.body)
		if err != nil {
			return nil, err
		}
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(data))
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
	glog.Info(string(resp.body))
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

// http://www.w3.org/TR/encoding
func (resp *HttpResponse) Encode(name string) *HttpResponse {
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
