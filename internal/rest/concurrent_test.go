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
