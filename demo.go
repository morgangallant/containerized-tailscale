package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
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

func tailnet() string {
	if n, ok := os.LookupEnv("TAILSCALE_TAILNET"); ok {
		return n
	}
	panic("missing TAILSCALE_TAILNET environment variable")
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

const privateUrl = "http://[fd7a:115c:a1e0:ab12:4843:cd96:6249:f919]/"

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

const (
	tailnetGetListDevicesUrl = "https://api.tailscale.com/api/v2/tailnet/%s/devices"
	tailnetDeviceDeleteUrl   = "https://api.tailscale.com/api/v2/device/%s"
)

type device struct {
	Hostname string `json:"hostname"`
	ID       string `json:"id"`
}

var errNotFound = errors.New("device with id not found")

func findDeviceIDWithHostname(tkey, tailnet, hn string) (string, error) {
	url := fmt.Sprintf(tailnetGetListDevicesUrl, tailnet)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(tkey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-ok status code %d returned from tailscale api: %s", resp.StatusCode, resp.Status)
	}
	var buf struct {
		Devices []device `json:"devices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&buf); err != nil {
		return "", err
	}
	for _, d := range buf.Devices {
		if d.Hostname == hn {
			return d.ID, nil
		}
	}
	return "", errNotFound
}

func removeMachineFromTailscale(tkey, tailnet, hn string) error {
	id, err := findDeviceIDWithHostname(tkey, tailnet, hn)
	if err == errNotFound {
		return nil
	} else if err != nil {
		return err
	}
	url := fmt.Sprintf(tailnetDeviceDeleteUrl, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(tkey, "")
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

func runServer(s *http.Server, hn, tkey, tailnet string, stop <-chan struct{}) error {
	errc := make(chan error)
	go func() {
		errc <- s.ListenAndServe()
	}()
	var err error
	select {
	case err = <-errc:
	case <-stop:
	}
	if err != nil && err != http.ErrServerClosed {
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
	if hn != "mbp" {
		if err := removeMachineFromTailscale(tkey, tailnet, hn); err != nil {
			return err
		}
		log.Printf("Removed ourselves from Tailscale.")
	} else {
		log.Printf("Skipping tailscale removal cuz we're in dev.")
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
	tkey, tnet := tailscaleAPIKey(), tailnet()
	rctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Printf("Got signal %s, shutting down.", sig.String())
		shutdown()
	}()

	return runServer(s, hn, tkey, tnet, rctx.Done())
}
