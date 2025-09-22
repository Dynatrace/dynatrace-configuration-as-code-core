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
	"encoding/json"
	"fmt"
)

type Metadata struct {
	ID                string  `json:"id"`
	ExternalID        string  `json:"externalId"`
	Actor             string  `json:"actor"`
	Owner             string  `json:"owner"`
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Version           int     `json:"version"`
	IsPrivate         bool    `json:"isPrivate"`
	OriginAppID       *string `json:"originAppId,omitempty"`
	OriginExtensionID *string `json:"originExtensionId,omitempty"`
}

func UnmarshallMetadata(b []byte) (Metadata, error) {
	var m Metadata
	if err := json.Unmarshal(b, &m); err != nil {
		return Metadata{}, fmt.Errorf("unable to unmarshal metadata: %w", err)
	}

	return m, nil
}
