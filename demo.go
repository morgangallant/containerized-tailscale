package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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

func tailscaleAPIKey() string {
	if k, ok := os.LookupEnv("TAILSCALE_API_KEY"); ok {
		return k
	}
	panic("missing TAILSCALE_API_KEY environment variable")
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

const privateUrl = "http://100.73.249.25/"

func handler(client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		req, err := http.NewRequestWithContext(r.Context(), "GET", privateUrl, nil)
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
		fmt.Fprintf(w, "Completed request to %s in %d ms.\n\n", privateUrl, time.Since(start).Milliseconds())
		fmt.Fprintln(w, "This is a demonstration of an outgoing SOCKS5 HTTP request from a container running inside Railway (https://railway.app) using some of the new Tailscale 1.16 Userspace Networking stuff.")
	}
}

const tailnetDeviceDeleteUrl = "https://api.tailscale.com/api/v2/device/%s"

func removeMachineFromTailscale(hn, tkey string) error {
	url := fmt.Sprintf(tailnetDeviceDeleteUrl, hn)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s:", tkey))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-ok status code %d returned from tailscale api: %s", resp.StatusCode, resp.Status)
	}
	return nil
}

func runServer(s *http.Server, hn, tkey string, stop <-chan struct{}) error {
	errc := make(chan error)
	go func() {
		errc <- s.ListenAndServe()
	}()
	var err error
	select {
	case err = <-errc:
	case <-stop:
	}
	if err != http.ErrServerClosed {
		return err
	}
	// At this point, we want to shutdown the HTTP server and cleanup.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		return err
	}
	// We only remove ourselves from Tailscale if we are in production.
	// For demo purposes, "production" means we're on a linux machine.
	if runtime.GOOS == "linux" {
		if err := removeMachineFromTailscale(hn, tkey); err != nil {
			return err
		}
	}
	return nil
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler(client))

	s := &http.Server{
		Addr:         ":" + port(),
		Handler:      mux,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second * 10,
	}

	// Setup signal listeners to automatically get rid of the server
	tkey := tailscaleAPIKey()
	rctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Printf("Got signal %s, shutting down.", sig.String())
		shutdown()
	}()

	return runServer(s, hn, tkey, rctx.Done())
}
