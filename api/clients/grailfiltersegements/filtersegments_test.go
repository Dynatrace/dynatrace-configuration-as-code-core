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
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/grailfiltersegements"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

func TestList(t *testing.T) {
	t.Run("add-fields are NOT specified", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())

			fields := r.URL.Query()["add-fields"]
			require.ElementsMatch(t, []string{"EXTERNALID"}, fields)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := grailfiltersegements.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.List(context.TODO(), rest.RequestOptions{})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("add-fields are specified by caller", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())

			fields := r.URL.Query()["add-fields"]
			require.ElementsMatch(t, []string{"user_defined"}, fields)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := grailfiltersegements.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.List(context.TODO(), rest.RequestOptions{QueryParams: url.Values{"add-fields": []string{"user_defined"}}})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGet(t *testing.T) {
	t.Run("add-fields are NOT specified", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())

			fields := r.URL.Query()["add-fields"]
			require.ElementsMatch(t, []string{"INCLUDES", "VARIABLES", "EXTERNALID", "RESOURCECONTEXT"}, fields)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := grailfiltersegements.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Get(context.TODO(), "id", rest.RequestOptions{})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("add-fields are specified by caller", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())

			fields := r.URL.Query()["add-fields"]
			require.ElementsMatch(t, []string{"user_defined"}, fields)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := grailfiltersegements.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Get(context.TODO(), "id", rest.RequestOptions{QueryParams: url.Values{"add-fields": []string{"user_defined"}}})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
