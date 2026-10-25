all: proxy.json bin/proxyscrape bin/vpngate

proxy.json: bin/proxyscrape
	touch proxy.json
	./bin/proxyscrape proxy.json

bin/%: %/main.go
	mkdir -p bin
	go build -o $@ $<
