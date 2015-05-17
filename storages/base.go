// Package storages provides various mechanism to store limit counters.
package storages

import (
	"time"
)

type ICounterStorage interface {
	IncrBy(string, int64, time.Duration)
	Get(string) (int64, bool)
}
