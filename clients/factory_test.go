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
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/accounts"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/automation"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/buckets"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/openpipeline"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/segments"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/slo"
)

const failedToParseURL = "failed to parse URL"

func TestClientCreation(t *testing.T) {
	f := Factory().
		WithPlatformURL("https://example.com/api").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithAccessToken("abc123").
		WithClassicURL("https://example.com/classicapi").
		WithAccountURL("https://example.com/accountapi").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &buckets.Client{}, clientInstance)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &segments.Client{}, clientInstance)

	clientInstance, err = f.SLOClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &slo.Client{}, clientInstance)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &automation.Client{}, clientInstance)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &documents.Client{}, clientInstance)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &openpipeline.Client{}, clientInstance)

	clientInstance, err = f.AccountClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &accounts.Client{}, clientInstance)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, restClient)

	restClient, err = f.CreateClassicClient()
	assert.NoError(t, err)
	assert.NotNil(t, restClient)
}

func TestClientMissingPlatformURL(t *testing.T) {
	f := Factory().
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithAccessToken("abc123").
		WithClassicURL("https://example.com/classicapi").
		WithAccountURL("https://example.com/accountapi").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrPlatformURLMissing)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrPlatformURLMissing)

	clientInstance, err = f.SLOClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrPlatformURLMissing)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrPlatformURLMissing)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrPlatformURLMissing)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrPlatformURLMissing)

	clientInstance, err = f.AccountClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &accounts.Client{}, clientInstance)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.Nil(t, restClient)
	assert.ErrorIs(t, err, ErrPlatformURLMissing)

	restClient, err = f.CreateClassicClient()
	assert.NoError(t, err)
	assert.NotNil(t, restClient)
}

func TestClientMissingOAuthCredentials(t *testing.T) {
	f := Factory().
		WithPlatformURL("https://example.com/api").
		WithAccessToken("abc123").
		WithClassicURL("https://example.com/classicapi").
		WithAccountURL("https://example.com/accountapi").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	clientInstance, err = f.SLOClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	clientInstance, err = f.AccountClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.Nil(t, restClient)
	assert.ErrorIs(t, err, ErrOAuthCredentialsMissing)

	restClient, err = f.CreateClassicClient()
	assert.NoError(t, err)
	assert.NotNil(t, restClient)
}

func TestClientPlatformURLParsingError(t *testing.T) {

	f := Factory().
		WithPlatformURL(":invalid-url").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithAccessToken("abc123").
		WithClassicURL("https://example.com/classicapi").
		WithAccountURL("https://example.com/accountapi").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}

	clientInstance, err := f.BucketClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, failedToParseURL)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, failedToParseURL)

	clientInstance, err = f.SLOClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, failedToParseURL)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, failedToParseURL)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, failedToParseURL)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, failedToParseURL)

	clientInstance, err = f.AccountClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &accounts.Client{}, clientInstance)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.Nil(t, restClient)
	assert.ErrorContains(t, err, failedToParseURL)

	restClient, err = f.CreateClassicClient()
	assert.NoError(t, err)
	assert.NotNil(t, restClient)
}

func TestClientMissingAccountURL(t *testing.T) {
	f := Factory().
		WithPlatformURL("https://example.com/api").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithAccessToken("abc123").
		WithClassicURL("https://example.com/classicapi").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &buckets.Client{}, clientInstance)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &segments.Client{}, clientInstance)

	clientInstance, err = f.SLOClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &slo.Client{}, clientInstance)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &automation.Client{}, clientInstance)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &documents.Client{}, clientInstance)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &openpipeline.Client{}, clientInstance)

	clientInstance, err = f.AccountClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorIs(t, err, ErrAccountURLMissing)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, restClient)

	restClient, err = f.CreateClassicClient()
	assert.NoError(t, err)
	assert.NotNil(t, restClient)
}

func TestClientAccountURLParsingError(t *testing.T) {
	f := Factory().
		WithPlatformURL("https://example.com/api").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithAccessToken("abc123").
		WithClassicURL("https://example.com/classicapi").
		WithAccountURL(":invalid-url").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &buckets.Client{}, clientInstance)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &segments.Client{}, clientInstance)

	clientInstance, err = f.SLOClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &slo.Client{}, clientInstance)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &automation.Client{}, clientInstance)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &documents.Client{}, clientInstance)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &openpipeline.Client{}, clientInstance)

	clientInstance, err = f.AccountClient(t.Context())
	assert.Nil(t, clientInstance)
	assert.ErrorContains(t, err, failedToParseURL)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, restClient)

	restClient, err = f.CreateClassicClient()
	assert.NoError(t, err)
	assert.NotNil(t, restClient)
}

func TestClientMissingClassicURL(t *testing.T) {
	f := Factory().
		WithPlatformURL("https://example.com/api").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithAccessToken("abc123").
		WithAccountURL("https://example.com/accountapi").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &buckets.Client{}, clientInstance)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &segments.Client{}, clientInstance)

	clientInstance, err = f.SLOClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &slo.Client{}, clientInstance)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &automation.Client{}, clientInstance)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &documents.Client{}, clientInstance)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &openpipeline.Client{}, clientInstance)

	clientInstance, err = f.AccountClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &accounts.Client{}, clientInstance)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, restClient)

	restClient, err = f.CreateClassicClient()
	assert.Nil(t, restClient)
	assert.ErrorIs(t, err, ErrClassicURLMissing)
}

func TestClientMissingAccessToken(t *testing.T) {
	f := Factory().
		WithPlatformURL("https://example.com/api").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithClassicURL("https://example.com/classicapi").
		WithAccountURL("https://example.com/accountapi").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &buckets.Client{}, clientInstance)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &segments.Client{}, clientInstance)

	clientInstance, err = f.SLOClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &slo.Client{}, clientInstance)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &automation.Client{}, clientInstance)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &documents.Client{}, clientInstance)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &openpipeline.Client{}, clientInstance)

	clientInstance, err = f.AccountClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &accounts.Client{}, clientInstance)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, restClient)

	restClient, err = f.CreateClassicClient()
	assert.Nil(t, restClient)
	assert.ErrorIs(t, err, ErrAccessTokenMissing)
}

func TestClientClassicURLParsingError(t *testing.T) {
	f := Factory().
		WithPlatformURL("https://example.com/api").
		WithOAuthCredentials(clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://auth.example.com/token",
		}).
		WithAccessToken("abc123").
		WithClassicURL(":invalid-url").
		WithAccountURL("https://example.com/accountapi").
		WithUserAgent("MyUserAgent")

	var clientInstance interface{}
	clientInstance, err := f.BucketClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &buckets.Client{}, clientInstance)

	clientInstance, err = f.SegmentsClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &segments.Client{}, clientInstance)

	clientInstance, err = f.SLOClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &slo.Client{}, clientInstance)

	clientInstance, err = f.AutomationClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &automation.Client{}, clientInstance)

	clientInstance, err = f.DocumentClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &documents.Client{}, clientInstance)

	clientInstance, err = f.OpenPipelineClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &openpipeline.Client{}, clientInstance)

	clientInstance, err = f.AccountClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, clientInstance)
	assert.IsType(t, &accounts.Client{}, clientInstance)

	restClient, err := f.CreatePlatformClient(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, restClient)

	restClient, err = f.CreateClassicClient()
	assert.Nil(t, restClient)
	assert.ErrorContains(t, err, failedToParseURL)
}
