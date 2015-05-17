[![GoDoc](https://godoc.org/github.com/didip/tollbooth?status.svg)](http://godoc.org/github.com/didip/tollbooth)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/didip/tollbooth/master/LICENSE)

## Tollbooth

This is a generic middleware to rate limit HTTP requests.


## Usage
```
package main

import (
    "github.com/didip/tollbooth"
    "github.com/didip/tollbooth/storages"
    "net/http"
    "time"
)

func HelloHandler(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("Hello, World!"))
}

func main() {
    // 1. Create a request limiter storage.
    storage := storages.NewInMemory()

    // 2. Create a request limiter per handler.
    http.Handle("/", tollbooth.LimitByIPFuncHandler(storage, tollbooth.NewRequestLimit(1, time.Second), HelloHandler))
    http.ListenAndServe(":12345", nil)
}
```
