package main

import (
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
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: URL parameter is required")
		fmt.Println("Usage: go run ratelimit.go -url <target-url>")
		os.Exit(1)
	}

	currentRate := 5            // Starting rate (requests per second)
	timeout := 15 * time.Second // Duration for each rate test
	cooldown := 10 * time.Second // Wait time between tests
	rateLimitFound := false

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	fmt.Printf("Starting rate limit detection for %s\n", *url)

	for !rateLimitFound {
		fmt.Printf("Testing rate: %d req/s for %v\n", currentRate, timeout)

		var wg sync.WaitGroup
		successfulRequests := 0
		failedRequests := 0
		rateLimitedRequests := 0
		otherErrors := 0

		startTime := time.Now()
		requests := make(chan bool, currentRate*int(timeout.Seconds()))

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
						resp, err := client.Get(*url)
						if err != nil {
							failedRequests++
							continue
						}
						defer resp.Body.Close()

						if resp.StatusCode == http.StatusOK {
							successfulRequests++
						} else if resp.StatusCode == 429 { // 429 Too Many Requests
							rateLimitedRequests++
						} else {
							otherErrors++
						}

						requests <- true
					case <-time.After(time.Until(startTime.Add(timeout))):
						return
					}
				}
			}()
		}

		// Wait for test duration or until all requests are completed
		select {
		case <-time.After(timeout):
		case <-requests:
		}

		wg.Wait()

		totalRequests := successfulRequests + rateLimitedRequests + otherErrors
		actualRate := float64(totalRequests) / timeout.Seconds()

		fmt.Printf("Results: %d total requests (%.1f req/s actual rate)\n", totalRequests, actualRate)
		fmt.Printf("  Success: %d, Rate Limited: %d, Other Errors: %d\n", 
			successfulRequests, rateLimitedRequests, otherErrors)

		// Check if we hit the rate limit
		if rateLimitedRequests > 0 {
			rateLimitFound = true
			fmt.Printf("\nRate limit found at approximately %d requests per second\n", currentRate)
			fmt.Printf("Detailed results:\n")
			fmt.Printf("- First rate limit detected at: %d req/s\n", currentRate)
			fmt.Printf("- Successful requests before limit: %d\n", successfulRequests)
			fmt.Printf("- Rate limited requests: %d\n", rateLimitedRequests)
			fmt.Printf("- Other errors: %d\n", otherErrors)
			fmt.Printf("- Actual achieved rate: %.1f req/s\n", actualRate)
		} else {
			fmt.Printf("No rate limit detected at %d req/s. Waiting %v before next test...\n\n", currentRate, cooldown)
			time.Sleep(cooldown)
			currentRate += 5 // Increase rate by 5 req/s for next test
		}
	}
}
