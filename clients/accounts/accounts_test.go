/*
 * @license
 * Copyright 2025 Dynatrace LLC
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

package accounts_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/accounts"
	accountmanagement "github.com/dynatrace/dynatrace-configuration-as-code-core/gen/account_management"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	responses := []testutils.ResponseDef{
		{
			GET: func(t *testing.T, req *http.Request) testutils.Response {
				return testutils.Response{
					ResponseCode: http.StatusOK,
					ResponseBody: `{"enabled": true, "allowWebhookOverride": true}`,
				}
			},
		},
	}

	server := testutils.NewHTTPTestServer(t, responses)
	defer server.Close()

	client := accounts.NewClient(rest.NewClient(server.URL(), server.Client()))

	data, resp, err := client.EnvironmentManagementAPI.GetConfig(t.Context(), "account-uuid", "env-uuid").Execute()
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, &accountmanagement.IpConfigDto{Enabled: true, AllowWebhookOverride: true, AdditionalProperties: map[string]any{}}, data)
}
