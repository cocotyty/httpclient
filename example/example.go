package main

import (
	"fmt"
	"io"
	"os"

	"github.com/cocotyty/httpclient"
	"github.com/cocotyty/httpclient/cache"
)

func main() {
	httpclient.
		Get("https://github.com/search").
		Query("utf8", "✓").
		Query("q", "httpclient").Dump().Send()

	httpclient.Get("http://baidu.com").DumpTo(nopCloser{os.Stderr}, nopCloser{os.Stderr}).Send()
	httpclient.Get("http://baidu.com").Dump().Send()
	builder := &httpclient.Builder{
		Cache: &cache.Cache{},
	}
	htmlContent, _ := builder.Request("sessionid").Get().Url("https://www.baidu.com").Send().Encoding("utf-8").String()
	fmt.Println(htmlContent)
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
