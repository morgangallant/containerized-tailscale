package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
	return &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}, nil
}

func run() error {
	client, err := socks5Client()
	if err != nil {
		return err
	}
	http.HandleFunc("/", handler(client))
	return http.ListenAndServe(":"+port(), nil)
}

const chuckNorrisAPIUrl = "https://api.chucknorris.io/jokes/random"

func handler(client *http.Client) http.HandlerFunc {
	type response struct {
		Value string `json:"value"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), "GET", chuckNorrisAPIUrl, nil)
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
		var rbody response
		if err := json.NewDecoder(resp.Body).Decode(&rbody); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Your Chuck Norris fact is: %s", rbody.Value)
	}
}
