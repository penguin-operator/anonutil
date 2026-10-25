# AnonUtil
A collection of custom utilities for internet anonymity.

## Programs

### [proxyspeed](proxyspeed/proxyspeed.go)
A Go library for testing proxy speeds. It measures how long it takes to make HTTP requests through a proxy and calculates the transfer speed.

**Usage as a library:**

```go
package main

import (
	"github.com/penguin-operator/anonutils/proxyspeed"
)

func main() {
	proxies := []string{
		"socks5://127.0.0.1:1080",
		"socks5://192.168.1.100:8080",
	}
	results := proxyspeed.TestProxies(proxies)
	for _, p := range results {
		fmt.Printf("%s: %d bytes/sec\n", p.Proxy, p.Speed)
	}
}
```

**Functions:**
- `TestProxy(proxy string) (Proxy, error)` - Tests a single proxy's speed
- `TestProxies(proxies []string) []Proxy` - Tests multiple proxies concurrently

---

### [proxyscrape](proxyscrape/main.go)
A command-line tool for scraping, testing, and ranking free proxies from proxyscrape.com. It downloads proxy lists, tests their speed, sorts them by performance, and saves results to a JSON file.

**Usage:**

```bash
# Test proxies from a file and save results
./proxyscrape output.json

# Use the fastest proxy to scrape more proxies
./proxyscrape -x output.json
```

**Options:**
- `-x` - Use the fastest proxy from the output file for scraping

**How it works:**
1. Loads existing proxy results from a JSON file (or creates new ones)
2. Tests all loaded proxies for speed
3. Sorts proxies by speed (fastest first)
4. Saves updated results to the same file
5. If `-x` flag is used, uses the fastest proxy to scrape additional proxies from proxyscrape.com
6. Tests the newly scraped proxies and saves all results

---

### [vpngate](vpngate/main.go)
A tool for scraping OpenVPN configurations from vpngate.net, decoding them, and saving them with geolocation information.

**Usage:**

```bash
# Save configs in custom format
./vpngate "{country}/{region}/{city}/{proto}/{ip}:{port}.ovpn"
```

**Options:**
- `-x` - Use a proxy file (JSON) to access the VPN gate service

**How it works:**
1. Scrapes VPN server configurations from vpngate.net
2. Decodes base64-encoded OpenVPN configs
3. Extracts protocol, IP address, and port from each config
4. Geolocates each IP using ip-api.com
5. Saves configs in the specified format with country, region, and city inserted
6. Creates directory structure automatically

**Supported format placeholders:**
- `{country}` - VPN server country (lowercase, underscores)
- `{region}` - VPN server region (lowercase, underscores)
- `{city}` - VPN server city (lowercase, underscores)
- `{proto}` - VPN protocol (tcp/udp)
- `{ip}` - VPN server IP address
- `{port}` - VPN server port

**Example:**
```bash
# Save configs to: US/New_York/New_York/tls-vpn-123.ovpn
./vpngate "{country}/{region}/{city}/tls-vpn-{port}.ovpn"
```
