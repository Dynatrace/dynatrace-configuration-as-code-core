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

package segments_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/segments"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestList(t *testing.T) {
	t.Run("check call", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments:lean", r.URL.Path)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := segments.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.List(t.Context(), rest.RequestOptions{})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("add-fields are NOT specified", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())

			fields := r.URL.Query()["add-fields"]
			require.ElementsMatch(t, []string{"EXTERNALID"}, fields)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := segments.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.List(t.Context(), rest.RequestOptions{})
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

		c := segments.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.List(t.Context(), rest.RequestOptions{QueryParams: url.Values{"add-fields": []string{"user_defined"}}})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGet(t *testing.T) {
	t.Run("check call", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", r.URL.Path)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := segments.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Get(t.Context(), "uid", rest.RequestOptions{})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Returns error for missing id", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "", rest.RequestOptions{})
		assert.Zero(t, resp)
		assert.ErrorContains(t, err, "id")
	})

	t.Run("add-fields are NOT specified", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())
			fields := r.URL.Query()["add-fields"]
			require.ElementsMatch(t, []string{"INCLUDES", "VARIABLES", "EXTERNALID", "RESOURCECONTEXT"}, fields)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := segments.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Get(t.Context(), "id", rest.RequestOptions{})
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

		c := segments.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Get(t.Context(), "id", rest.RequestOptions{QueryParams: url.Values{"add-fields": []string{"user_defined"}}})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.String())
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments", r.URL.Path)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	c := segments.NewClient(rest.NewClient(u, server.Client()))

	resp, err := c.Create(t.Context(), []byte{}, rest.RequestOptions{})
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUpdate(t *testing.T) {

	t.Run("add owner and uid", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", r.URL.Path)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := segments.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Update(t.Context(), "uid", []byte{}, rest.RequestOptions{})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Returns error for missing id", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "", []byte{}, rest.RequestOptions{})
		assert.Zero(t, resp)
		assert.ErrorContains(t, err, "id")
	})
}

func TestDelete(t *testing.T) {
	t.Run("check call", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", r.URL.Path)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := segments.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Delete(t.Context(), "uid", rest.RequestOptions{})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Returns error for missing id", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, "", rest.RequestOptions{})
		assert.Zero(t, resp)
		assert.ErrorContains(t, err, "id")
	})
}
