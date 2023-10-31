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
	"golang.org/x/oauth2/clientcredentials"
	"net/url"
	"time"
)

// ErrOAuthCredentialsMissing indicates that no OAuth2 client credentials were provided.
var ErrOAuthCredentialsMissing = errors.New("no OAuth2 client credentials provided")

// ErrEnvironmentURLMissing indicates that no environment URL was provided.
var ErrEnvironmentURLMissing = errors.New("no environment URL provided")

// Factory creates a factory-like component that is used to create API client instances.
func Factory() factory {
	return factory{}
}

// factory represents a factory-like component for creating API client instances.
type factory struct {
	url          string                    // The base URL of the API
	oauthConfig  *clientcredentials.Config // Configuration for OAuth2 client credentials
	userAgent    string                    // The User-Agent header to be set
	httpListener *rest.HTTPListener        // The HTTP listener to be set
}

// WithOAuthCredentials sets the OAuth2 client credentials configuration for the factory.
func (f factory) WithOAuthCredentials(config clientcredentials.Config) factory {
	f.oauthConfig = &config
	return f
}

// WithEnvironmentURL sets the base URL for the API.
func (f factory) WithEnvironmentURL(u string) factory {
	f.url = u
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

// AccountClient creates and reaturns a new instance of accounts.Client for interacting with the accounts API.
func (f factory) AccountClient(accountManagementURL string) (*accounts.Client, error) {
	restClient, err := f.createClientForAccount(accountManagementURL)
	if err != nil {
		return nil, err
	}
	return accounts.NewClient(restClient), nil
}

// AutomationClient creates and returns a new instance of automation.Client for interacting with the automation API.
func (f factory) AutomationClient() (*automation.Client, error) {
	restClient, err := f.createClient()
	if err != nil {
		return nil, err
	}
	return automation.NewClient(restClient), nil
}

// BucketClient creates and returns a new instance of buckets.Client for interacting with the bucket API.
func (f factory) BucketClient() (*buckets.Client, error) {
	restClient, err := f.createClient()
	if err != nil {
		return nil, err
	}
	return buckets.NewClient(restClient), nil
}

// BucketClientWithRetrySettings creates and returns a new instance of buckets.Client with non-default retry settings.
// For details about how retry settings are used, see buckets.WithRetrySettings.
func (f factory) BucketClientWithRetrySettings(maxRetries int, durationBetweenTries time.Duration, maxWaitDuration time.Duration) (*buckets.Client, error) {
	restClient, err := f.createClient()
	if err != nil {
		return nil, err
	}
	return buckets.NewClient(restClient, buckets.WithRetrySettings(maxRetries, durationBetweenTries, maxWaitDuration)), nil
}

func (f factory) createClientForAccount(accManagementURL string) (*rest.Client, error) {
	if f.oauthConfig == nil {
		return nil, ErrOAuthCredentialsMissing
	}

	parsedURL, err := url.Parse(accManagementURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %w", f.url, err)
	}

	restClient := rest.NewClient(parsedURL, auth.NewOAuthBasedClient(context.TODO(), *f.oauthConfig), rest.WithHTTPListener(f.httpListener))
	if f.userAgent != "" {
		restClient.SetHeader("User-Agent", f.userAgent)
	}

	return restClient, nil
}
func (f factory) createClient() (*rest.Client, error) {
	if f.url == "" {
		return nil, ErrEnvironmentURLMissing
	}

	if f.oauthConfig == nil {
		return nil, ErrOAuthCredentialsMissing
	}

	parsedURL, err := url.Parse(f.url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %w", f.url, err)
	}

	restClient := rest.NewClient(parsedURL, auth.NewOAuthBasedClient(context.TODO(), *f.oauthConfig), rest.WithHTTPListener(f.httpListener))
	if f.userAgent != "" {
		restClient.SetHeader("User-Agent", f.userAgent)
	}

	return restClient, nil
}
