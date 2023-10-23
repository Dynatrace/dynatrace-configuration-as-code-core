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
Package api groups packages simplifying Dynatrace API access.

Notably it contains the following packages
  - [github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest] containing an extended rest client with optional features like rate-limiting, request logging, etc.
  - [github.com/dynatrace/dynatrace-configuration-as-code-core/api/auth] containing methods for creating authenticated rest clients
  - [github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients] containing client implementations for specific Dynatrace APIs
*/
package api
