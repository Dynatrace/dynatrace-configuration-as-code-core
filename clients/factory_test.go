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

package clients

import (
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/automation"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/buckets"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"
	"testing"
)

func TestClientCreation(t *testing.T) {

	// Prepare a factory instance with necessary configurations
	f := Factory().
		WithEnvironmentURL("https://example.com/api").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient()
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &buckets.Client{}, clientInstance)

	clientInstance, err = f.AutomationClient()
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &automation.Client{}, clientInstance)

	//... other clients
}

func TestClientMissingEnvironmentURL(t *testing.T) {

	// Prepare a factory instance without an environment URL
	f := Factory().
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient()
	assert.Error(t, err)
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrEnvironmentURLMissing)

	clientInstance, err = f.AutomationClient()
	assert.Error(t, err)
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrEnvironmentURLMissing)

	//... other clients
}

func TestClientMissingOAuthCredentials(t *testing.T) {

	// Prepare a factory instance without OAuth credentials
	f := Factory().
		WithEnvironmentURL("https://example.com/api").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient()
	assert.Error(t, err)
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	clientInstance, err = f.AutomationClient()
	assert.Error(t, err)
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	//... other clients
}

func TestClientURLParsingError(t *testing.T) {

	// Prepare a factory instance with a malformed URL
	f := Factory().
		WithEnvironmentURL(":invalid-url").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient()
	assert.Error(t, err)
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, "failed to parse URL")

	clientInstance, err = f.AutomationClient()
	assert.Error(t, err)
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, "failed to parse URL")

	//... other clients
}
