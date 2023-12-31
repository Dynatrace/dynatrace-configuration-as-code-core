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

package accounts

import (
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/gen/account_management"
)

type Client struct {
	*accountmanagement.APIClient
}

// NewClient creates a new accounts.Client to interact with the accounts API
func NewClient(cl *rest.Client) *Client {
	config := &accountmanagement.Configuration{
		Servers: accountmanagement.ServerConfigurations{{
			URL: cl.BaseURL().String(),
		}},
		HTTPClient: cl,
	}
	return &Client{APIClient: accountmanagement.NewAPIClient(config)}
}
