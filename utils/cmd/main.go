package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/deepissue/core/utils"
)

func main() {
	url, _ := url.Parse("http://127.0.0.1:8000/openapi.json")
	req := http.Request{
		URL:    url,
		Method: "GET",
	}
	for i := 0; i < 10; i++ {
		utils.PerformHTTPRequest(&req, 1)
	}
	for {
		time.Sleep(1)
	}
}
