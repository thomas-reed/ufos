package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	url := flag.String("url", "", "health check URL")
	flag.Parse()

	target := *url
	if target == "" {
		target = os.Getenv("PROBE_URL")
	}
	if target == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		target = "http://127.0.0.1:" + port + "/healthz"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "unexpected status: %s\n", res.Status)
		os.Exit(1)
	}
}