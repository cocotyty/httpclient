package httpclient

import (
	"bytes"
	"net/url"

	"golang.org/x/text/encoding"
)

func encodeQuery(source [][]string, encoding encoding.Encoding) []byte {
	buf := bytes.NewBuffer(nil)
	for _, kv := range source {
		k, v := kv[0], kv[1]
		buf.WriteString(k)
		buf.WriteByte('=')
		if encoding != nil {
			v, _ = encoding.NewEncoder().String(v)
		}
		buf.WriteString(url.QueryEscape(v))
		buf.WriteByte('&')
	}
	result := buf.Bytes()
	if result[len(result)-1] == '&' {
		result = result[:len(result)-1]
	}
	return result
}

func encodeForm(source map[string][]string, encoding encoding.Encoding) []byte {
	buf := bytes.NewBuffer(nil)
	for k, strs := range source {
		for _, v := range strs {
			buf.WriteString(k)
			buf.WriteByte('=')
			if encoding!=nil {
				v, _ = encoding.NewEncoder().String(v)
			}
			buf.WriteString(url.QueryEscape(v))
			buf.WriteByte('&')
		}
	}
	result := buf.Bytes()
	if result[len(result)-1] == '&' {
		result = result[:len(result)-1]
	}
	return result
}
