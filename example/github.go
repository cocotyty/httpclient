package main

import (
	"github.com/cocotyty/httpclient"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/PuerkitoBio/goquery"
	"time"
	"bytes"
)

func main() {
	fmt.Println(
		httpclient.
		Get("https://github.com/search").
		Query("utf8", "✓").
		Query("q", "httpclient").
		Send().
		String())

	hs := &httpclient.HttpService{Cache:cache.New(1 * time.Minute, 1 * time.Minute)}

	user := "user1"
	page, err := hs.Get(user).Url("http://github.com/login").Send().Body()
	if err != nil {
		panic(err)
	}

	dom, _ := goquery.NewDocumentFromReader(bytes.NewBuffer(page))
	authenticityToken, _ := dom.Find("input[name=authenticity_token]").Attr("value")
	fmt.Println(
		hs.Post(user).
		Url("https://github.com/session").
		Param("commit", "Sign in").
		Param("utf8", "✓").
		Param("authenticity_token", authenticityToken).
		Param("login", "").
		Param("password", "").
		Send().
		String())
}
