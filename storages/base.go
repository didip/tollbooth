package storages

import (
	"time"
)

type ICounterStorage interface {
	IncrBy(string, int64, time.Duration)
	Get(string) (int64, bool)
}
