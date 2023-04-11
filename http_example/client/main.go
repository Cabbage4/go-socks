package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

func main() {
	proxyUrl, err := url.Parse("socks5://127.0.0.1:18888")
	if err != nil {
		panic(err)
	}

	mc := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	rsp, err := mc.Get("http://127.0.0.1:8081")
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}

	log.Printf("%s\n", b)
}
