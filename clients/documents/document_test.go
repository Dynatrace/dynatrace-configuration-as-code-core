/*
 * @license
 * Copyright 2026 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package documents

import (
	"bytes"
	"io"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func toPtr[T any](v T) *T { return &v }

func TestWriteDocument(t *testing.T) {
	tests := []struct {
		name        string
		meta        Metadata
		content     []byte
		wantValues  map[string][]string // form fields that must be present with these values
		wantAbsent  []string            // form fields that must not be present
		wantContent []byte              // expected content part (nil = no content part expected)
	}{
		{
			name: "only required fields - all optionals skipped",
			meta: Metadata{Type: Dashboard, Name: "my-doc", IsPrivate: true},
			wantValues: map[string][]string{
				"type":      {"dashboard"},
				"name":      {"my-doc"},
				"isPrivate": {"true"},
			},
			wantAbsent: []string{"id", "description", "labels", "isReshareable"},
		},
		{
			name: "all settable fields set",
			meta: Metadata{
				Type:          Notebook,
				Name:          "doc",
				ID:            "the-id",
				IsPrivate:     false,
				Description:   toPtr("a description"),
				Labels:        []string{"alpha", "beta"},
				IsReshareable: toPtr(true),
			},
			content: []byte("file-content"),
			wantValues: map[string][]string{
				"type":          {"notebook"},
				"name":          {"doc"},
				"isPrivate":     {"false"},
				"id":            {"the-id"},
				"description":   {"a description"},
				"labels":        {"alpha", "beta"},
				"isReshareable": {"true"},
			},
			wantContent: []byte("file-content"),
		},
		{
			name: "non-nil empty labels clears via a single empty field",
			meta: Metadata{Type: Dashboard, Name: "n", Labels: []string{}, IsReshareable: toPtr(false)},
			wantValues: map[string][]string{
				"labels":        {""},
				"isReshareable": {"false"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			_, boundary, err := writeDocument(body, tt.meta, tt.content)
			require.NoError(t, err)

			form, err := multipart.NewReader(body, boundary).ReadForm(1 << 20)
			require.NoError(t, err)
			t.Cleanup(func() { _ = form.RemoveAll() })

			for field, want := range tt.wantValues {
				assert.Equal(t, want, form.Value[field], "field %q", field)
			}
			for _, field := range tt.wantAbsent {
				assert.NotContains(t, form.Value, field, "field %q should be absent", field)
			}

			if tt.wantContent == nil {
				assert.Empty(t, form.File["content"], "no content part expected")
				return
			}
			require.Len(t, form.File["content"], 1)
			f, err := form.File["content"][0].Open()
			require.NoError(t, err)
			defer f.Close()
			got, err := io.ReadAll(f)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, got)
		})
	}
}

// failingWriter always fails, used to exercise writeDocument's error path.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestWriteDocument_WriteError(t *testing.T) {
	_, _, err := writeDocument(failingWriter{}, Metadata{Type: Dashboard, Name: "n"}, nil)
	assert.Error(t, err)
}
