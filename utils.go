package httpclient

import (
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"net/url"
)

func buildQueryEncoded(source [][]string, gb18030 bool) []byte {
	buf := bytes.NewBuffer(nil)
	for _, kv := range source {
		k, v := kv[0], kv[1]
		buf.WriteString(k)
		buf.WriteByte('=')
		if gb18030 {
			v, _ = simplifiedchinese.GB18030.NewEncoder().String(v)
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
	if result[len(result)-1] == '&' {
		result = result[:len(result)-1]
	}
	return result
}
