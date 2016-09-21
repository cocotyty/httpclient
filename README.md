# HTTP Client

eg. search github
```go
httpclient.
		Get("https://github.com/search").
		Query("utf8", "✓").
		Query("q", "httpclient").
		Send().
		String()
```

see example/github.go

