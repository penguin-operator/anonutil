package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/penguin-operator/anonutils/proxyspeed"
	"golang.org/x/net/proxy"
)

const (
	IP_API = "http://ip-api.com/json/"
	VPNGATE = "https://www.vpngate.net/api/iphone/"
)

func Filter[T any](s []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range s {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

func GetVPN(config string) (proto string, ip string, port string) {
	config = strings.ReplaceAll(config, "\r", "")
	lines := strings.Split(config, "\n")
	lines = Filter(lines, func(s string) bool {
		return !strings.HasPrefix(s, "#") && !strings.HasPrefix(s, ";") && s != ""
	})
	config = strings.Join(lines, "\n")
	config = strings.Trim(config, "\n")
	config = strings.Split(config, "<ca>")[0]
	lines = strings.Split(config, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "proto") {
			proto = strings.Split(line, " ")[1]
		}
		if strings.HasPrefix(line, "remote") {
			cfg := strings.Split(line, " ")
			ip, port = cfg[1], cfg[2]
		}
	}
	return
}

func GetIPGeo(ip string) (country, region, city string, err error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(IP_API + ip)
	for err != nil || resp.StatusCode != http.StatusOK {
		time.Sleep(60 * time.Second)
		resp, err = client.Get(IP_API + ip)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil { return }
	if body["country"] == nil || body["region"] == nil || body["city"] == nil {
		err = errors.New("malformed body")
		return
	}
	country, region, city = body["country"].(string), body["region"].(string), body["city"].(string)
	country, region, city = strings.ToLower(country), strings.ToLower(region), strings.ToLower(city)
	country, region, city = strings.ReplaceAll(country, " ", "_"), strings.ReplaceAll(region, " ", "_"), strings.ReplaceAll(city, " ", "_")
	return
}

func ProcessVPN(config []byte, format string) (err error) {
	proto, ip, port := GetVPN(string(config))
	country, region, city, err := GetIPGeo(ip)
	if err != nil { return }
	format = strings.ReplaceAll(format, "{country}", country)
	format = strings.ReplaceAll(format, "{region}", region)
	format = strings.ReplaceAll(format, "{city}", city)
	format = strings.ReplaceAll(format, "{proto}", proto)
	format = strings.ReplaceAll(format, "{ip}", ip)
	format = strings.ReplaceAll(format, "{port}", port)
	path := strings.Split(format, "/")
	for i := range path[:len(path)-1] {
		os.Mkdir(strings.Join(path[:i+1], "/"), 0755)
	}
	os.WriteFile(format, config, 0644)
	return
}

func main() {
	proxies_file := flag.String("x", "", "use proxies when specified a json file with proxies data")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		panic("no output file specified")
	}
	format := args[0]
	client := &http.Client{
		Timeout: 150 * time.Second,
	}
	if *proxies_file != "" {
		input, err := os.ReadFile(*proxies_file)
		if err != nil { panic(err) }
		var proxies []proxyspeed.Proxy
		json.Unmarshal(input, &proxies)
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
	fmt.Print("scraping vpngate ... ")
	resp, err := client.Get(VPNGATE)
	if err != nil { panic(err) }
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil { panic(err) }
	fmt.Println("done")
	data = bytes.ReplaceAll(data, []byte{'\r'}, []byte{})
	data = bytes.TrimLeft(data, "*vpn_servers\n#")
	data = bytes.TrimRight(data, "*\n")
	parser := csv.NewReader(bytes.NewReader(data))
	vpns, err := parser.ReadAll()
	if err != nil { panic(err) }
	fmt.Print("decoding servers configs ... ")
	var wg sync.WaitGroup
	var n int
	for _, vpn := range vpns[1:] {
		wg.Go(func () {
			config, err := base64.StdEncoding.DecodeString(vpn[14])
			if err != nil { panic(err) }
			if err = ProcessVPN(config, format); err == nil {
				n += 1
				fmt.Print("\rdecoding servers configs (", strconv.Itoa(n), ") ... ")
			}
		})
	}
	wg.Wait()
	fmt.Println("done")
}
