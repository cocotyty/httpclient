package httpclient

import (
	"bytes"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/htmlindex"
)

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

var contentTypeMatchCharset = regexp.MustCompile(`\bcharset=([\w|\-]*)`)

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
		} else {
			selector = doc.Find(`meta[http-equiv="Content-Type"]`)
			if selector != nil && selector.Size() > 0 {
				attrContent, exists := selector.Attr("content")
				if exists {
					subMatch := contentTypeMatchCharset.FindStringSubmatch(attrContent)
					if len(subMatch) == 2 {
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
