/*
 * @license
 * Copyright 2025 Dynatrace LLC
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
	"encoding/json"
	"fmt"
	"time"
)

// Metadata mirrors the document metadata returned by the API. It is also the
// input to Client.Create and Client.Update.
//
// Settable fields (accepted by the API on create and update):
// ID, Name, Type, IsPrivate, Description, IsReshareable, Labels.
//
// Read-only fields (populated by the API; ignored on create and update):
// Owner, Version, ModificationInfo, Access, OriginAppID, OriginExtensionID.
type Metadata struct {
	// Settable fields.
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	IsPrivate     bool     `json:"isPrivate"`
	Description   *string  `json:"description,omitempty"`
	IsReshareable *bool    `json:"isReshareable,omitempty"`
	Labels        []string `json:"labels,omitempty"`

	// Read-only fields — populated by the API, ignored on create and update.
	ModificationInfo  *ModificationInfo `json:"modificationInfo,omitempty"`
	Version           int               `json:"version"`
	Owner             string            `json:"owner"`
	OriginAppID       *string           `json:"originAppId,omitempty"`
	OriginExtensionID *string           `json:"originExtensionId,omitempty"`
	Access            []string          `json:"access,omitempty"`
}

// ModificationInfo captures the read-only audit information the API returns for
// a document.
type ModificationInfo struct {
	CreatedBy          string    `json:"createdBy"`
	CreatedTime        time.Time `json:"createdTime"`
	LastModifiedBy     string    `json:"lastModifiedBy"`
	LastModifiedTime   time.Time `json:"lastModifiedTime"`
	LastModifyingAppID *string   `json:"lastModifyingAppId"`
}

func UnmarshallMetadata(b []byte) (Metadata, error) {
	var m Metadata
	if err := json.Unmarshal(b, &m); err != nil {
		return Metadata{}, fmt.Errorf("unable to unmarshal metadata: %w", err)
	}

	return m, nil
}
