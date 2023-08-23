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

package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/api/auth"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/api/bucketclient"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/rest"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2/clientcredentials"
)

// ErrOAuthCredentialsMissing is returned when no OAuth2 client credentials are provided.
var ErrOAuthCredentialsMissing = errors.New("no OAuth2 client credentials provided")

// ErrEnvironmentURLMissing is returned when no URL to an environment is provided
var ErrEnvironmentURLMissing = errors.New("no environment URL provided")

// ClientFactory is a factory for creating API client instances.
var ClientFactory = clientFactory{}

// clientFactory represents a factory for creating API client instances.
type clientFactory struct {
	url         string                    // The base URL of the API.
	oauthConfig *clientcredentials.Config // Configuration for OAuth2 client credentials.
	logger      logr.Logger               // A logging interface.
}

// WithOAuthCredentials sets the OAuth2 client credentials configuration for the factory.
func (f clientFactory) WithOAuthCredentials(config clientcredentials.Config) clientFactory {
	f.oauthConfig = &config
	return f
}

// WithEnvironmentURL sets the base URL for the API.
func (f clientFactory) WithEnvironmentURL(u string) clientFactory {
	f.url = u
	return f
}

// WithLogger sets the logger interface for the factory.
func (f clientFactory) WithLogger(logger logr.Logger) clientFactory {
	f.logger = logger
	return f
}

// BucketClient creates and returns a new instance of bucketclient.Client for interacting with the bucket API
func (f clientFactory) BucketClient() (*bucketclient.Client, error) {
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

	return bucketclient.New(rest.NewClient(parsedURL, auth.NewOAuthBasedClient(context.TODO(), *f.oauthConfig), f.logger), f.logger), nil
}
