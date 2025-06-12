/*
 * @license
 * Copyright 2023 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// OAuthMockServer creates an OAuth server that just returns a valid OAuth access token response
func OAuthMockServer(t *testing.T, accessToken string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		assert.NoError(t, err)
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))

		resp := map[string]interface{}{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600, // seconds
		}
		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(resp)
		assert.NoError(t, err)
	}))
}
