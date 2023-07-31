package rest

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestConcurrentRequestLimiter_AcquireAndRelease(t *testing.T) {
	limiter := NewConcurrentRequestLimiter(2)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		limiter.Acquire()
		wg.Done()
	}()
	go func() {
		limiter.Acquire()
		wg.Done()
	}()
	wg.Wait()
	wg.Add(2)
	go func() {
		limiter.Acquire()
		wg.Done()
	}()
	go func() {
		limiter.Release()
		wg.Done()
	}()
	wg.Wait()
}

func TestConcurrentRequestLimiter_AcquireWithoutLimit(t *testing.T) {
	limiter := NewConcurrentRequestLimiter(0)
	limiter.Acquire()
	limiter.Acquire()
	limiter.Acquire()
	assert.Nil(t, limiter.sem, "Semaphore should be nil when maxConcurrent is 0")
	limiter = NewConcurrentRequestLimiter(-1)
	limiter.Acquire()
	limiter.Acquire()
	limiter.Acquire()
	assert.Nil(t, limiter.sem, "Semaphore should be nil when maxConcurrent is -1")
}
