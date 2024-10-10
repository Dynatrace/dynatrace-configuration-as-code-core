// @license
// Copyright 2023 Dynatrace LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	limitHeader = "X-RateLimit-Limit"
	resetHeader = "X-RateLimit-Reset"

	defaultTimeout = time.Millisecond * 100
)

type RateLimiter struct {
	Lock  sync.RWMutex
	Clock Clock

	limiter      *rate.Limiter
	resetAt      *time.Time
	resetTimeout *time.Duration
}

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type realtimeClock struct{}

func (realtimeClock) Now() time.Time {
	return time.Now()
}

func (realtimeClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		Clock: realtimeClock{},
	}
}

func (rl *RateLimiter) Update(ctx context.Context, status int, headers http.Header) {
	rl.Lock.Lock()
	defer rl.Lock.Unlock()

	logger := logr.FromContextOrDiscard(ctx)

	limit, err := extractLimit(headers)
	if err == nil && limit > 0 {
		if rl.limiter == nil {
			logger.V(1).Info(fmt.Sprintf("Rate limit set based on HTTP response. Client limited to %v requests per second", limit), "limit", limit)
			rl.limiter = rate.NewLimiter(limit, 1)
		} else if limit != rl.limiter.Limit() {
			logger.V(1).Info(fmt.Sprintf("Rate limit updated based on HTTP response. Client limited to %v requests per second", limit), "limit", limit)
			rl.limiter.SetLimit(limit)
		}
	}

	if status != http.StatusTooManyRequests {
		// no hard limit reached at the moment, carry on
		rl.resetAt = nil
		rl.resetTimeout = nil
		return
	}

	reset, err := extractTimeout(headers)
	if err != nil {
		logger.V(1).Info(fmt.Sprintf("Received 429 TooManyRequests HTTP response but using default timeout of %s because could not extract one: %s", defaultTimeout, err.Error()), "timeout", defaultTimeout, "headers", headers)
		rt := defaultTimeout
		rl.resetTimeout = &rt
		ra := rl.Clock.Now().Add(defaultTimeout)
		rl.resetAt = &ra
		return
	}

	rl.resetAt = &reset
	rt := reset.Sub(rl.Clock.Now())
	rl.resetTimeout = &rt

	logger.V(1).Info(fmt.Sprintf("Received 429 TooManyRequests HTTP. Timeout: %s", rl.resetTimeout), "timeout", rl.resetTimeout)

}

// extractLimit tries to parse the limitHeader into a rate.Limit for use with the soft-limit rate.Limiter.
func extractLimit(header http.Header) (rate.Limit, error) {
	lstr := header.Get(limitHeader)
	limit, err := strconv.Atoi(lstr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse header %q with value %q: %w", limitHeader, lstr, err)
	}
	return rate.Limit(limit), nil
}

// extractTimeout tries to parse the resetHeader into time.Time used for the 429 API response hard limit.
// As Dynatrace APIs don't return a time delta, but a Unix timestamp this will return a reset timestamp in server time.
func extractTimeout(header http.Header) (time.Time, error) {
	rstr := header.Get(resetHeader)
	reset, err := strconv.ParseInt(rstr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse header %q with value %q: %w", resetHeader, rstr, err)
	}

	return time.Unix(reset, 0), nil
}

// Wait blocks in case a hard API limit was reached, or the request/second limit was exceeded.
// In case of a hard limit, the method will block until after its reset time is reached.
// In case of the soft request/second limit it will block until requests are available again.
// See rate.Limiter for details on how soft-limits works.
func (rl *RateLimiter) Wait(ctx context.Context) {
	rl.Lock.RLock()
	defer rl.Lock.RUnlock()

	logger := logr.FromContextOrDiscard(ctx)

	if rl.resetAt != nil && rl.resetTimeout != nil && rl.resetAt.After(rl.Clock.Now()) {
		// hard limit triggered via 429 API response, wait until its timeout is reached
		<-rl.Clock.After(*rl.resetTimeout)
	}

	if rl.limiter != nil {
		err := rl.limiter.Wait(ctx)
		if err != nil {
			logger.Error(err, "Client-side rate limiting failed")
		}
	}
}
