package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("WAKAPI_PORT")
	if port == "" {
		port = "3000"
	}

	url := fmt.Sprintf("http://127.0.0.1:%s/api/health", port)

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Healthcheck failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Healthcheck returned status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	os.Exit(0)
}
