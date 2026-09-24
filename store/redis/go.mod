module github.com/didip/tollbooth/v8/store/redis

go 1.24

require (
	github.com/alicebob/miniredis/v2 v2.39.0
	github.com/didip/tollbooth/v8 v8.0.0
	github.com/redis/go-redis/v9 v9.22.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace github.com/didip/tollbooth/v8 => ../../
