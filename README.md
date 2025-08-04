# cashbill.go

[![Go Reference](https://pkg.go.dev/badge/github.com/IAmRadek/cashbill.go.svg)](https://pkg.go.dev/github.com/IAmRadek/cashbill.go)

Go implementation of Cashbill API

## Cashbill Documentation

https://api.cashbill.pl/api/payment-gateway/cashbill-payments-api

## Usage

```go
package main

import (
    "net/http"
    "time"

    "github.com/IAmRadek/cashbill.go"
)

func main() {
    // Create a new API client with default settings
    api := cashbill.NewAPI("your-shop-id", "your-secret")

    // Create a new API client with a custom HTTP client
    customClient := &http.Client{
        Timeout: 10 * time.Second,
    }
    apiWithCustomClient := cashbill.NewAPI("your-shop-id", "your-secret", cashbill.WithHTTPClient(customClient))

    // Create a new test API client
    testApi := cashbill.NewTestAPI("your-shop-id", "your-secret")

    // Create a new test API client with a custom HTTP client
    testApiWithCustomClient := cashbill.NewTestAPI("your-shop-id", "your-secret", cashbill.WithHTTPClient(customClient))

    // Use the API clients...
}
```
