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
	"context"
	"errors"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/auth"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/accounts"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/automation"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/buckets"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/openpipeline"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/url"
	"time"
)

// ErrOAuthCredentialsMissing indicates that no OAuth2 client credentials were provided.
var ErrOAuthCredentialsMissing = errors.New("no OAuth2 client credentials provided")

// ErrPlatformURLMissing indicates that no platform API URL was provided.
var ErrPlatformURLMissing = errors.New("no platform API URL provided")

// ErrClassicURLMissing indicates that no classic API URL was provided.
var ErrClassicURLMissing = errors.New("no classic API URL provided")

// ErrAccountURLMissing indicates that no account API URL was provided.
var ErrAccountURLMissing = errors.New("no account API URL provided")

// ErrAccessTokenMissing indicates that no access token was provided.
var ErrAccessTokenMissing = errors.New("no access token provided")

// Factory creates a factory-like component that is used to create API client instances.
func Factory() factory {
	return factory{}
}

// factory represents a factory-like component for creating API client instances.
type factory struct {
	platformURL            string                    // The base URL for platform APIs
	classicURL             string                    // The base URL for classic APIs
	accountURL             string                    // The base URL for account APIs
	oauthConfig            *clientcredentials.Config // Configuration for OAuth2 client credentials
	accessToken            string                    // Access token for API
	userAgent              string                    // The User-Agent header to be set
	httpListener           *rest.HTTPListener        // The HTTP listener to be set
	concurrentRequestLimit int                       // The number of allowed concurrent requests
	rateLimiterEnabled     bool                      // Enables rate limiter for clients
	retryOptions           *rest.RetryOptions        // The retry strategy
}

// WithOAuthCredentials sets the OAuth2 client credentials configuration for the factory.
func (f factory) WithOAuthCredentials(config clientcredentials.Config) factory {
	f.oauthConfig = &config
	return f
}

// WithAccessToken sets the access token for the factory.
func (f factory) WithAccessToken(accessToken string) factory {
	f.accessToken = accessToken
	return f
}

// WithPlatformURL sets the base URL for accessing the platform API.
func (f factory) WithPlatformURL(u string) factory {
	f.platformURL = u
	return f
}

// WithClassicURL sets the base URL for accessing the classic API.
func (f factory) WithClassicURL(u string) factory {
	f.classicURL = u
	return f
}

// WithAccountURL sets the base URL for accessing the account API.
func (f factory) WithAccountURL(u string) factory {
	f.accountURL = u
	return f
}

// WithUserAgent sets the User-Agent header.
func (f factory) WithUserAgent(userAgent string) factory {
	f.userAgent = userAgent
	return f
}

// WithHTTPListener sets the given HTTPListener to be used by the
// underlying rest/http client
func (f factory) WithHTTPListener(listener *rest.HTTPListener) factory {
	f.httpListener = listener
	return f
}

// WithConcurrentRequestLimit sets the given request limit that specifies how many
// requests can be triggered concurrently by the underlying rest/http client.
func (f factory) WithConcurrentRequestLimit(limit int) factory {
	f.concurrentRequestLimit = limit
	return f
}

// WithRateLimiter enables a RateLimiter for Clients.
func (f factory) WithRateLimiter(enabled bool) factory {
	f.rateLimiterEnabled = enabled
	return f
}

// WithRetryOptions sets the RetryOptions for the underlying rest/http clients.
func (f factory) WithRetryOptions(retryOptions *rest.RetryOptions) factory {
	f.retryOptions = retryOptions
	return f
}

// AccountClient creates and reaturns a new instance of accounts.Client for interacting with the accounts API.
func (f factory) AccountClient() (*accounts.Client, error) {
	restClient, err := f.createAccountClient()
	if err != nil {
		return nil, err
	}
	return accounts.NewClient(restClient), nil
}

// AutomationClient creates and returns a new instance of automation.Client for interacting with the automation API.
func (f factory) AutomationClient() (*automation.Client, error) {
	restClient, err := f.CreatePlatformClient()
	if err != nil {
		return nil, err
	}
	return automation.NewClient(restClient), nil
}

// BucketClient creates and returns a new instance of buckets.Client for interacting with the bucket API.
func (f factory) BucketClient() (*buckets.Client, error) {
	restClient, err := f.CreatePlatformClient()
	if err != nil {
		return nil, err
	}
	return buckets.NewClient(restClient), nil
}

// DocumentClient creates and returns a new instance of documents.Client for interacting with the document API.
func (f factory) DocumentClient() (*documents.Client, error) {
	restClient, err := f.CreatePlatformClient()
	if err != nil {
		return nil, err
	}
	return documents.NewClient(restClient), nil
}

// BucketClientWithRetrySettings creates and returns a new instance of buckets.Client with non-default retry settings.
// For details about how retry settings are used, see buckets.WithRetrySettings.
func (f factory) BucketClientWithRetrySettings(maxRetries int, durationBetweenTries time.Duration, maxWaitDuration time.Duration) (*buckets.Client, error) {
	restClient, err := f.CreatePlatformClient()
	if err != nil {
		return nil, err
	}
	return buckets.NewClient(restClient, buckets.WithRetrySettings(maxRetries, durationBetweenTries, maxWaitDuration)), nil
}

// OpenPipelineClient creates and returns a new instance of openpipeline.Client for interacting with the openPipeline API.
func (f factory) OpenPipelineClient() (*openpipeline.Client, error) {
	restClient, err := f.CreatePlatformClient()
	if err != nil {
		return nil, err
	}
	return openpipeline.NewClient(restClient), nil
}

func (f factory) createAccountClient() (*rest.Client, error) {
	if f.oauthConfig == nil {
		return nil, ErrOAuthCredentialsMissing
	}

	if f.accountURL == "" {
		return nil, ErrAccountURLMissing
	}

	return f.createRestClient(f.accountURL, auth.NewOAuthBasedClient(context.TODO(), *f.oauthConfig))
}

// CreatePlatformClient creates a REST client configured for accessing platform APIs.
func (f factory) CreatePlatformClient() (*rest.Client, error) {
	if f.oauthConfig == nil {
		return nil, ErrOAuthCredentialsMissing
	}

	if f.platformURL == "" {
		return nil, ErrPlatformURLMissing
	}

	return f.createRestClient(f.platformURL, auth.NewOAuthBasedClient(context.TODO(), *f.oauthConfig))
}

// CreateClassicClient creates a REST client configured for accessing classic APIs.
func (f factory) CreateClassicClient() (*rest.Client, error) {
	if f.accessToken == "" {
		return nil, ErrAccessTokenMissing
	}

	if f.classicURL == "" {
		return nil, ErrClassicURLMissing
	}

	return f.createRestClient(f.classicURL, auth.NewTokenBasedClient(f.accessToken))
}

func (f factory) createRestClient(u string, httpClient *http.Client) (*rest.Client, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %w", u, err)
	}

	opts := []rest.Option{rest.WithHTTPListener(f.httpListener)}
	if f.rateLimiterEnabled {
		opts = append(opts, rest.WithRateLimiter())
	}

	if f.retryOptions != nil {
		opts = append(opts, rest.WithRetryOptions(f.retryOptions))
	}

	restClient := rest.NewClient(parsedURL, httpClient, opts...)
	if f.userAgent != "" {
		restClient.SetHeader("User-Agent", f.userAgent)
	}

	return restClient, nil
}
