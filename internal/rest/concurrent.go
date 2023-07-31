package rest

// ConcurrentRequestLimiter represents a component for limiting concurrent requests.
type ConcurrentRequestLimiter struct {
	sem chan struct{}
}

// NewConcurrentRequestLimiter creates a new instance of ConcurrentRequestLimiter with the specified limit to apply.
func NewConcurrentRequestLimiter(maxConcurrent int) *ConcurrentRequestLimiter {
	if maxConcurrent <= 0 {
		// If maxConcurrent is 0 or negative, set the semaphore channel to nil to represent no limit.
		return &ConcurrentRequestLimiter{
			sem: nil,
		}
	}
	return &ConcurrentRequestLimiter{
		sem: make(chan struct{}, maxConcurrent),
	}
}

// Acquire acquires a slot from the concurrent request limiter to check for maximum concurrent requests.
func (c *ConcurrentRequestLimiter) Acquire() {
	if c.sem != nil {
		c.sem <- struct{}{}
	}
}

// Release releases a slot from the concurrent request limiter to allow subsequent requests to proceed.
func (c *ConcurrentRequestLimiter) Release() {
	if c.sem != nil {
		select {
		case <-c.sem:
		default:
			// If the semaphore is already empty, do nothing
		}
	}
}
