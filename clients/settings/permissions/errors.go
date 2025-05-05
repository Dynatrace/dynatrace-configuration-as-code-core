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

package permissions

import (
	"errors"
	"fmt"
)

var (
	ErrorMissingObjectID     = errors.New("objectID cannot be empty")
	ErrorMissingAccessorID   = errors.New("accessorID cannot be empty")
	ErrorMissingAccessorType = errors.New("accessorType cannot be empty")
)

type ErrorPermissionGet struct {
	Err error
}

func (e ErrorPermissionGet) Error() string {
	return fmt.Sprintf("failed to get permission(s): %v", e.Err)
}

type ErrorPermissionCreate struct {
	Err error
}

func (e ErrorPermissionCreate) Error() string {
	return fmt.Sprintf("failed to create permission: %v", e.Err)
}

type ErrorPermissionUpdate struct {
	Err error
}

func (e ErrorPermissionUpdate) Error() string {
	return fmt.Sprintf("failed to update permission: %v", e.Err)
}

type ErrorPermissionDelete struct {
	Err error
}

func (e ErrorPermissionDelete) Error() string {
	return fmt.Sprintf("failed to delete permission object: %v", e.Err)
}
