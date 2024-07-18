/*
 * @license
 * Copyright 2023 Dynatrace LLC
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
	"context"
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

//go:generate mockgen -source client.go -package=documents -destination=client_mock.go
type client interface {
	Get(ctx context.Context, id string) (*http.Response, error)
	Create(ctx context.Context, doc documents.Document) (*http.Response, error)
	Patch(ctx context.Context, id string, version int, doc documents.Document) (*http.Response, error)
	List(ctx context.Context, requestOptions rest.RequestOptions) (*http.Response, error)
	Update(ctx context.Context, id string, data []byte, requestOptions rest.RequestOptions) (*http.Response, error)
	Delete(ctx context.Context, id string, requestOptions rest.RequestOptions) (*http.Response, error)

	Trash(ctx context.Context, id string) (*http.Response, error)
}

var _ client = (*documents.Client)(nil)
