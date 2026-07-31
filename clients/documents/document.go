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
	"io"
	"mime/multipart"
	"strconv"
)

// writeDocument serializes the settable fields of meta plus content into a
// multipart body. The following server-managed Metadata fields are read-only
// and are ignored here: Owner, Version, ModificationInfo, Access, OriginAppID,
// and OriginExtensionID.
func writeDocument(w io.Writer, meta Metadata, content []byte) (*multipart.Writer, error) {
	writer := multipart.NewWriter(w)

	if err := writer.WriteField("type", meta.Type); err != nil {
		return nil, err
	}
	if err := writer.WriteField("name", meta.Name); err != nil {
		return nil, err
	}
	if err := writer.WriteField("isPrivate", strconv.FormatBool(meta.IsPrivate)); err != nil {
		return nil, err
	}
	if meta.ID != "" {
		if err := writer.WriteField("id", meta.ID); err != nil {
			return nil, err
		}
	}
	if meta.Description != nil {
		if err := writer.WriteField("description", *meta.Description); err != nil {
			return nil, err
		}
	}
	// Labels are sent as one repeated "labels" field per value. A nil slice
	// leaves the existing labels untouched; a non-nil but empty slice clears
	// them by sending a single empty "labels" field.
	if meta.Labels != nil {
		if len(meta.Labels) == 0 {
			if err := writer.WriteField("labels", ""); err != nil {
				return nil, err
			}
		}
		for _, label := range meta.Labels {
			if err := writer.WriteField("labels", label); err != nil {
				return nil, err
			}
		}
	}
	if meta.IsReshareable != nil {
		if err := writer.WriteField("isReshareable", strconv.FormatBool(*meta.IsReshareable)); err != nil {
			return nil, err
		}
	}
	if content != nil {
		part, err := writer.CreateFormFile("content", meta.Name)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(content); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return writer, err
	}

	return writer, nil
}
