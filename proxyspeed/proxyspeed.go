package proxyspeed

import (
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

const (
	IP_API = "http://ip-api.com/json"
)

type Proxy struct {
	Proxy string `json:"proxy"`
	Speed int64  `json:"speed"`
}

func TestProxy(p string) (Proxy, error) {
	url, err := url.Parse(p)
	if err != nil { return Proxy{}, err }
	dialer, err := proxy.FromURL(url, proxy.Direct)
	if err != nil { return Proxy{}, err }
	transport := &http.Transport{
		Dial: dialer.Dial,
	}
	client := &http.Client{
		Transport: transport,
		Timeout: 10 * time.Second,
	}
	start := time.Now()
	resp, err := client.Get(IP_API)
	if err != nil { return Proxy{}, err }
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil { return Proxy{}, err }
	elapsed := time.Since(start)
	speed := float64(len(content)) / elapsed.Seconds()
	return Proxy { p, int64(speed) }, nil
}

func TestProxies(proxies []string) []Proxy {
	var res []Proxy
	var wg sync.WaitGroup
	for _, p := range proxies {
		wg.Go(func() {
			if proxy, err := TestProxy(p); err == nil && proxy.Speed != 0 {
				res = append(res, proxy)
			}
		})
	}
	wg.Wait()
	return res
}
