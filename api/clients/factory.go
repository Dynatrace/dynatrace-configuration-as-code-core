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
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/buckets"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"net/url"

	"golang.org/x/oauth2/clientcredentials"
)

// ErrOAuthCredentialsMissing is returned when no OAuth2 client credentials are provided.
var ErrOAuthCredentialsMissing = errors.New("no OAuth2 client credentials provided")

// ErrEnvironmentURLMissing is returned when no URL to an environment is provided
var ErrEnvironmentURLMissing = errors.New("no environment URL provided")

// Factory creates a factory component to be used
// to create API client instances
func Factory() factory {
	return factory{}
}

// factory represents a factory for creating API client instances.
type factory struct {
	url         string                    // The base URL of the API.
	oauthConfig *clientcredentials.Config // Configuration for OAuth2 client credentials.
	userAgent   string                    // The User-Agent header to be set
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

// WithUserAgent sets the user agent
func (f factory) WithUserAgent(userAgent string) factory {
	f.userAgent = userAgent
	return f
}

// BucketClient creates and returns a new instance of buckets.Client for interacting with the bucket API
func (f factory) BucketClient() (*buckets.Client, error) {
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

	restClient := rest.NewClient(parsedURL, auth.NewOAuthBasedClient(context.TODO(), *f.oauthConfig))
	if f.userAgent != "" {
		restClient.SetHeader("User-Agent", f.userAgent)
	}

	return buckets.NewClient(restClient), nil
}
