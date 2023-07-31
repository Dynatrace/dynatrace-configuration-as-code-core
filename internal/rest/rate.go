package rest

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	Lock         sync.Mutex
	Limit        int
	ResetAt      time.Time
	ResetTimeout time.Duration
}

// TODO: implement simple rate limiter based on the X-RateLimit-Reset header
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (rl *RateLimiter) Update(headers http.Header) {
	rl.Lock.Lock()
	defer rl.Lock.Unlock()
	// TODO: implement
}

func (rl *RateLimiter) Allow() bool {
	//TODO: implement
	return true
}
