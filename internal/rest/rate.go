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
