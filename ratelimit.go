package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	// Parse command-line arguments
	url := flag.String("url", "", "Target URL to test for rate limiting")
	method := flag.String("method", "GET", "HTTP method to use: GET or POST")
	postData := flag.String("data", "", "POST data to send (used only if method is POST)")
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: URL parameter is required")
		fmt.Println("Usage: go run ratelimit.go -url <target-url> [-method GET|POST] [-data 'key=value']")
		os.Exit(1)
	}

	currentRate := 5
	timeout := 15 * time.Second
	cooldown := 10 * time.Second
	rateLimitFound := false

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	fmt.Printf("Starting rate limit detection for %s using %s\n", *url, *method)

	for !rateLimitFound {
		fmt.Printf("Testing rate: %d req/s for %v\n", currentRate, timeout)

		var wg sync.WaitGroup
		var mu sync.Mutex
		successfulRequests := 0
		rateLimitedRequests := 0
		otherErrors := 0

		startTime := time.Now()

		// Launch goroutines to make requests
		for i := 0; i < currentRate; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						var req *http.Request
						var err error

						if *method == "POST" {
							req, err = http.NewRequest("POST", *url, bytes.NewBuffer([]byte(*postData)))
							req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
						} else {
							req, err = http.NewRequest("GET", *url, nil)
						}

						if err != nil {
							mu.Lock()
							otherErrors++
							mu.Unlock()
							continue
						}

						resp, err := client.Do(req)
						if err != nil {
							mu.Lock()
							otherErrors++
							mu.Unlock()
							continue
						}
						resp.Body.Close()

						mu.Lock()
						switch resp.StatusCode {
						case http.StatusOK:
							successfulRequests++
						case http.StatusTooManyRequests:
							rateLimitedRequests++
						default:
							otherErrors++
						}
						mu.Unlock()

					case <-time.After(time.Until(startTime.Add(timeout))):
						return
					}
				}
			}()
		}

		wg.Wait()

		totalRequests := successfulRequests + rateLimitedRequests + otherErrors
		actualRate := float64(totalRequests) / timeout.Seconds()

		fmt.Printf("Results: %d total requests (%.1f req/s actual rate)\n", totalRequests, actualRate)
		fmt.Printf("  Success: %d, Rate Limited: %d, Other Errors: %d\n",
			successfulRequests, rateLimitedRequests, otherErrors)

		if rateLimitedRequests > 0 {
			rateLimitFound = true
			fmt.Printf("\nRate limit found at approximately %d requests per second\n", currentRate)
		} else {
			fmt.Printf("No rate limit detected at %d req/s. Waiting %v before next test...\n\n", currentRate, cooldown)
			time.Sleep(cooldown)
			currentRate += 5
		}
	}
}

