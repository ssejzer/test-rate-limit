# Rate Limit Detector

A Go script that automatically discovers the rate limits of a web server by incrementally increasing request rates until the limit is detected.

## Features

- Automatically tests increasing request rates (starting at 5 req/s)
- Detects HTTP 429 (Too Many Requests) responses
- Provides detailed reporting of rate limit thresholds
- Configurable test duration and cooldown periods
- Command-line interface for easy targeting of different endpoints

## Usage

```bash
go run ratelimit.go -url <target-url>
```

## Example

```bash
go run ratelimit.go -url https://api.example.com/v1/endpoint
```

## Expected Output

```text
Starting rate limit detection for https://api.example.com/v1/endpoint
Testing rate: 5 req/s for 15s
Results: 75 total requests (5.0 req/s actual rate)
  Success: 75, Rate Limited: 0, Other Errors: 0
No rate limit detected at 5 req/s. Waiting 10s before next test...

Testing rate: 10 req/s for 15s
Results: 150 total requests (10.0 req/s actual rate)
  Success: 150, Rate Limited: 0, Other Errors: 0
No rate limit detected at 10 req/s. Waiting 10s before next test...

Testing rate: 15 req/s for 15s
Results: 225 total requests (15.0 req/s actual rate)
  Success: 200, Rate Limited: 25, Other Errors: 0

Rate limit found at approximately 15 requests per second
Detailed results:
- First rate limit detected at: 15 req/s
- Successful requests before limit: 200
- Rate limited requests: 25
- Other errors: 0
- Actual achieved rate: 15.0 req/s
```

## Configuration Options

The script uses sensible defaults but can be modified by editing these constants in the source:

- Starting rate: 5 requests/second
- Test duration: 15 seconds per rate
- Cooldown period: 10 seconds between tests
- Rate increment: +5 requests/second each test

## How It Works

- Starts with 5 requests per second for 15 seconds
- If no rate limit is detected (no 429 responses):
  - Waits 10 seconds
  - Increases rate by 5 req/s
  - Repeats test
- When rate limit is detected:
  - Prints detailed report
  - Exits

## Best Practices

- Use against test/staging environments when possible
- Start with conservative rates for production systems
- Consider adding authentication headers if testing protected endpoints
- Be aware of any terms of service for the API you're testing

