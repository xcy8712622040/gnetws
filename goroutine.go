package gnetws

import (
	"github.com/panjf2000/ants/v2"
	"runtime"
	"sync"
	"time"
)

var (
	// Nonblocking decides what to do when submitting a new task to a full worker pool: waiting for a available worker
	// or returning nil directly.
	Nonblocking = true

	// ExpiryDuration is the interval time to clean up those expired workers.
	ExpiryDuration = 10 * time.Second

	// DefaultAntsPoolSize sets up the capacity of worker pool, 1024 * cpu.
	DefaultAntsPoolSize = (1 << 10) * runtime.NumCPU() * 2
)

type Pool = *ants.Pool
type PoolWithFunc = *ants.PoolWithFunc

var GoroutinePool = func() func() Pool {
	var pool Pool
	once := sync.Once{}
	options := ants.Options{
		Nonblocking:    Nonblocking,
		ExpiryDuration: ExpiryDuration,
	}

	return func() Pool {
		once.Do(func() {
			pool, _ = ants.NewPool(
				DefaultAntsPoolSize, ants.WithOptions(options),
			)
		})
		return pool
	}
}()
