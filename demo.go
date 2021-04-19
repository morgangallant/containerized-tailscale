package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/proxy"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func port() string {
	if p, ok := os.LookupEnv("PORT"); ok {
		return p
	}
	return "8080"
}

const proxyAddr = "localhost:1080"

func socks5Client() (*http.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}
	dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.Dial(network, address)
	}
	transport := &http.Transport{
		DialContext:       dialContext,
		DisableKeepAlives: true,
	}
	return &http.Client{Transport: transport, Timeout: 0}, nil
}

func run() error {
	hn, err := os.Hostname()
	if err != nil {
		return err
	}
	log.Printf("Running on hostname %s.", hn)
	client, err := socks5Client()
	if err != nil {
		return err
	}
	http.HandleFunc("/", handler(client))
	return http.ListenAndServe(":"+port(), nil)
}

const privateUrl = "http://100.73.249.25/"

func handler(client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		req, err := http.NewRequest("GET", privateUrl, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		fmt.Fprintf(w, "Completed request to http://100.73.249.25/ in %d ms.\n\n", time.Since(start).Milliseconds())
		fmt.Fprintln(w, "This is a demonstration of an outgoing SOCKS5 HTTP request from a container running inside Railway (https://railway.app) using some of the new Tailscale 1.16 Userspace Networking stuff.")
	}
}
