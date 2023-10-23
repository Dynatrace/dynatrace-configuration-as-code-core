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
Package clients contains 'smart' clients for the Dynatrace API.

These clients are generally based on those found in package api/clients, but implement logic to ensure the Dynatrace API
can be used for configuration-as-code use-cases reliably.

For the underlying clients see package [github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients].

In general, whereas the CRUD api/clients make a single API call per operation, the ones in this package make several and
react to API responses as needed.

For example, the clients in this package will resolve and follow pagination in their List methods, whereas an api/clients
implementation requires/allows the user to handle pagination and make several requests on their own.

Clients implement the following methods:
  - Get
  - List
  - Create
  - Update
  - Upsert (Create or Update as needed)
  - Delete

A clients.Factory simplifies creation of clients.
*/
package clients
