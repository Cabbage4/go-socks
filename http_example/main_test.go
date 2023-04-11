package main

import (
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestIpv4(t *testing.T) {
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

	if string(b) != "hello world" {
		t.Fail()
	}
}

func TestIpv6(t *testing.T) {
	proxyUrl, err := url.Parse("socks5://127.0.0.1:18888")
	if err != nil {
		panic(err)
	}

	mc := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	rsp, err := mc.Get("http://[::7f00:0001]:8081")
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}

	if string(b) != "hello world" {
		t.Fail()
	}
}

func TestDomain(t *testing.T) {
	proxyUrl, err := url.Parse("socks5://127.0.0.1:18888")
	if err != nil {
		panic(err)
	}

	mc := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	rsp, err := mc.Get("http://localhost:8081")
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}

	if string(b) != "hello world" {
		t.Fail()
	}
}

func TestUsernameAndPassword(t *testing.T) {
	proxyUrl, err := url.Parse("socks5://admin:admin@127.0.0.1:18888")
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

	if string(b) != "hello world" {
		t.Fail()
	}
}
