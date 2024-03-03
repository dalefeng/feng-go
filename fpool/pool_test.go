package fpool

import (
	"math"
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	_ = 1 << (10 * iota)
	Kib
	Mib
)

const (
	Param    = 100
	PoolSize = 1000
	TestSize = 10000
	n        = 1000000
)

var curMem uint64

const (
	RunTimes           = 1000000
	BenchParam         = 10
	DefaultExpiredTime = 10 * time.Second
)

func DemoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			DemoFunc()
			wg.Done()
		}()
	}

	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/Mib - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestHasPool(t *testing.T) {
	pool, _ := NewPool(math.MaxInt32)
	defer pool.Release()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = pool.Submit(func() {
			DemoFunc()
			wg.Done()
		})
	}
	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/Mib - curMem
	t.Logf("memory usage:%d MB\n", curMem)
	t.Logf("running worker :%d\n", pool.Running())
	t.Logf("free worker:%d\n", pool.Free())
}
