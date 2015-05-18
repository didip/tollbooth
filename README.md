[![GoDoc](https://godoc.org/github.com/didip/tollbooth?status.svg)](http://godoc.org/github.com/didip/tollbooth)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/didip/tollbooth/master/LICENSE)

## Tollbooth

This is a generic middleware to rate-limit HTTP requests.


## Five Minutes Tutorial
```
package main

import (
    "github.com/didip/tollbooth"
    "github.com/didip/tollbooth/storages/memory"
    "net/http"
    "time"
)

func HelloHandler(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("Hello, World!"))
}

func main() {
    // 1. Create a request limiter storage.
    storage := memory.New()

    // 2. Create a request limiter per handler.
    http.Handle("/", tollbooth.LimitFuncHandler(storage, tollbooth.NewLimiter(1, time.Second), HelloHandler))
    http.ListenAndServe(":12345", nil)
}
```

## Features

1. Rate-limit by remote IP, request path, request methods, custom headers, & basic auth usernames.
    ```
    limiter := tollbooth.NewLimiter(1, time.Second)

    // Limit only GET and POST requests.
    limiter.Methods = []string{"GET", "POST"}

    // Limit request headers containing certain values.
    // Usually, you prefetched these values from the database.
    limiter.Headers = make(map[string][]string)
    limiter.Headers["X-Access-Token"] = []string{"abc123", "xyz098"}

    // Limit based on certain basic auth usernames.
    // Usually, you prefetched these values from the database.
    limiter.BasicAuthUsers = []string{"bob", "joe", "didip"}
    ```

2. Each request handler can be rate-limited individually.

3. Compose your own middlware by using `LimitByKeyParts()`.

4. Customizable storage by implementing `ICounterStorage` interface.
