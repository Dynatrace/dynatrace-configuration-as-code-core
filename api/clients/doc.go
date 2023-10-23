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
Package clients contains simple CRUD clients for the Dynatrace API.

For 'smarter' clients with extended functionality see package [github.com/dynatrace/dynatrace-configuration-as-code-core/clients].

In general, the CRUD clients make a single API call per operation, but in some cases may make several if the Dynatrace API requires it.

Clients implement the following methods:

  - Get
  - List
  - Create
  - Update
  - Delete
*/
package clients
