all: bin/proxyscrape bin/vpngate

bin/%: %/main.go
	mkdir -p bin
	go build -o $@ $<
