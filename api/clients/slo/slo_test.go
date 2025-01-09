// @license
// Copyright 2025 Dynatrace LLC
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

package slo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/slo"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.String())
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/platform/slo/v1/slos", r.URL.Path)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	c := slo.NewClient(rest.NewClient(u, server.Client()))

	resp, err := c.List(context.TODO(), rest.RequestOptions{})
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.String())
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/platform/slo/v1/slos/uid", r.URL.Path)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	c := slo.NewClient(rest.NewClient(u, server.Client()))

	resp, err := c.Get(context.TODO(), "uid", rest.RequestOptions{})
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.String())
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/platform/slo/v1/slos", r.URL.Path)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	c := slo.NewClient(rest.NewClient(u, server.Client()))

	resp, err := c.Create(context.TODO(), []byte{}, rest.RequestOptions{})
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUpdate(t *testing.T) {
	t.Run("check call", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/slo/v1/slos/uid", r.URL.Path)
			require.Equal(t, "versionID", r.URL.Query().Get("optimistic-locking-version"))
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := slo.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Update(context.TODO(), "uid", "versionID", []byte{}, rest.RequestOptions{})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Returns error for missing id", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "", "versionID", []byte{}, rest.RequestOptions{})
		assert.Zero(t, resp)
		assert.ErrorContains(t, err, "id")
	})

	t.Run("Returns error for missing optimisticLockingVersion", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "ID", "", []byte{}, rest.RequestOptions{})
		assert.Zero(t, resp)
		assert.ErrorContains(t, err, "optimisticLockingVersion")
	})

}

func TestDelete(t *testing.T) {
	t.Run("check call", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.String())
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/slo/v1/slos/uid", r.URL.Path)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		c := slo.NewClient(rest.NewClient(u, server.Client()))

		resp, err := c.Delete(context.TODO(), "uid", rest.RequestOptions{})
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Returns error for missing id", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, "", rest.RequestOptions{})
		assert.Zero(t, resp)
		assert.ErrorContains(t, err, "id")
	})
}
