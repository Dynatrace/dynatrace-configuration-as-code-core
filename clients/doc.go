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

/*
Package clients contains clients for the Dynatrace API.

The clients in this package are specifically designed to enable reliable configuration-as-code use cases with the Dynatrace API.

Unlike standard CRUD clients that perform a single API call per operation, these clients execute multiple calls and dynamically adapt to API responses.
For example, they include logic to handle tasks like resolving and following pagination in their `List` methods.

Clients implement the following methods:
  - Get
  - List
  - Create
  - Update
  - Delete

A clients.Factory simplifies creation of clients.
*/
package clients
