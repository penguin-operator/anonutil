package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/penguin-operator/anonutils/proxyspeed"
	"golang.org/x/net/proxy"
)

const (
	PROXYSCRAPE = "https://api.proxyscrape.com/v4/free-proxy-list/get?request=display_proxies&proxy_format=protocolipport&format=text&protocol=socks5&timeout=10000"
)

type Proxies []proxyspeed.Proxy

func (p Proxies) Len() int {
	return len(p)
}

func (p Proxies) Less(i, j int) bool {
	return p[i].Speed < p[j].Speed
}

func (p Proxies) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func WriteResults(results []proxyspeed.Proxy, file string) error {
	sort.Sort(Proxies(results))
	output, err := json.MarshalIndent(results, "", "\t")
	if err != nil { return err }
	output = append(output, '\n')
	if err = os.WriteFile(file, output, 0644); err != nil { return err }
	return nil
}

func main() {
	use_proxy := flag.Bool("x", false,  "use fastest proxy from the output file")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		panic("no output file specified")
	}
	file := args[0]
	input, err := os.ReadFile(file)
	if err != nil { panic(err) }
	var proxies []proxyspeed.Proxy
	json.Unmarshal(input, &proxies)
	var purl []string
	for _, p := range proxies {
		purl = append(purl, p.Proxy)
	}
	fmt.Print("scanning proxies at ", file, " ... ")
	proxies = proxyspeed.TestProxies(purl)
	if err = WriteResults(proxies, file); err !=  nil { panic(err) }
	fmt.Println("done")
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	if *use_proxy {
		url, err := url.Parse(proxies[len(proxies)-1].Proxy)
		if err != nil { panic(err) }
		fmt.Println("using proxy", url)
		dialer, err := proxy.FromURL(url, proxy.Direct)
		if err != nil { panic(err) }
		transport := &http.Transport{
			Dial: dialer.Dial,
		}
		client.Transport = transport
	}
	fmt.Print("scraping proxyscrape ... ")
	resp, err := client.Get(PROXYSCRAPE)
	if err != nil { panic(err) }
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil { panic(err) }
	fmt.Println("done")
	for u := range strings.SplitSeq(string(data), "\r\n") {
		if !slices.Contains(purl, u) {
			purl = append(purl, u)
		}
	}
	fmt.Print("scanning gathered proxies ... ")
	proxies = proxyspeed.TestProxies(purl)
	if err = WriteResults(proxies, file); err !=  nil { panic(err) }
	fmt.Println("done")
}
