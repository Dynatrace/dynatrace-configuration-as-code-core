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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResponse_DecodeJSON(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		data := []byte(`{"key": "value"}`)
		response := Response{Data: data}
		type objType map[string]string

		obj, err := DecodeJSON[objType](response)

		assert.NoError(t, err)
		assert.Equal(t, "value", obj["key"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		data := []byte(`invalid-json`)
		response := Response{Data: data}

		type objType map[string]string

		_, err := DecodeJSON[objType](response)

		assert.Error(t, err)
		assert.Equal(t, "failed to unmarshal JSON: invalid character 'i' looking for beginning of value", err.Error())
	})
}
func TestListResponse_DecodeJSON(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		data := []byte(`{"results": [ { "key": "one" }, { "key": "two" } ] }`)
		response := ListResponse{
			Response: Response{
				Data: data,
			},
		}
		type objType map[string][]map[string]string

		obj, err := DecodeJSON[objType](response.Response)

		assert.NoError(t, err)
		assert.Len(t, obj["results"], 2)
		assert.Equal(t, "one", obj["results"][0]["key"])
		assert.Equal(t, "two", obj["results"][1]["key"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		data := []byte(`invalid-json`)
		response := ListResponse{
			Response: Response{
				Data: data,
			},
		}
		type objType map[string][]map[string]string

		_, err := DecodeJSON[objType](response.Response)

		assert.Error(t, err)
		assert.Equal(t, "failed to unmarshal JSON: invalid character 'i' looking for beginning of value", err.Error())
	})
}

func TestDecodeJSONObjects(t *testing.T) {
	response := ListResponse{
		Objects: [][]byte{
			[]byte(`{ "key": "one" }`),
			[]byte(`{ "key": "two" }`),
			[]byte(`{ "key": "three" }`),
		},
	}
	type val struct {
		Key string `json:"key"`
	}

	res, err := DecodeJSONObjects[val](response)

	assert.NoError(t, err)
	assert.Len(t, res, 3)
	assert.Equal(t, "one", res[0].Key)
	assert.Equal(t, "two", res[1].Key)
	assert.Equal(t, "three", res[2].Key)
}
