//go:build live

// @license
// Copyright 2024 Dynatrace LLC
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

package grailfiltersegements_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/grailfiltersegements"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/clientcredentials"
)

func Test_List(t *testing.T) {
	config := clientcredentials.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		TokenURL:     os.Getenv("TOKEN_URL"),
	}

	u, _ := url.Parse(os.Getenv("BASE_URL"))
	client := grailfiltersegements.NewClient(
		rest.NewClient(u, config.Client(context.TODO())))

	r, err := client.List(context.TODO())

	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.NotEmpty(t, r.Body)
}

func Test_Create(t *testing.T) {
	config := clientcredentials.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		TokenURL:     os.Getenv("TOKEN_URL"),
	}

	u, _ := url.Parse(os.Getenv("BASE_URL"))
	client := grailfiltersegements.NewClient(
		rest.NewClient(u, config.Client(context.TODO())))

	payload := []byte(`{
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false
}'`)
	r, err := client.Create(context.TODO(), payload)

	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, r.StatusCode)
	assert.NotEmpty(t, r.Body)
}

func Test_Get(t *testing.T) {
	config := clientcredentials.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		TokenURL:     os.Getenv("TOKEN_URL"),
	}

	u, _ := url.Parse(os.Getenv("BASE_URL"))
	client := grailfiltersegements.NewClient(
		rest.NewClient(u, config.Client(context.TODO())))

	r, err := client.Get(context.TODO(), "QElQbQcjq3S", rest.RequestOptions{})

	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.NotEmpty(t, r.Body)
}
