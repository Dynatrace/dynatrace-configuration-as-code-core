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

package rest_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

func TestShouldRetry(t *testing.T) {
	testcases := []struct {
		statusCode int
		want       bool
	}{
		{
			statusCode: http.StatusOK,
			want:       false,
		},
		{
			statusCode: http.StatusForbidden,
			want:       false,
		},
		{
			statusCode: http.StatusNotFound,
			want:       true,
		},
		{
			statusCode: http.StatusBadRequest,
			want:       true,
		},
	}

	for _, tt := range testcases {
		t.Run(fmt.Sprintf("Should be %v if status is %d", tt.want, tt.statusCode), func(t *testing.T) {
			got := rest.ShouldRetry(tt.statusCode)
			assert.Equal(t, tt.want, got)
		})
	}
}
