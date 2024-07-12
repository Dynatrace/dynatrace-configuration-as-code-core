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

type Document struct {
	Kind       string
	Name       string
	ExternalID string
	Public     bool
	Content    []byte
}

func (d *Document) write(w io.Writer) (*multipart.Writer, error) {
	writer := multipart.NewWriter(w)

	if err := writer.WriteField("type", d.Kind); err != nil {
		return nil, err
	}
	if err := writer.WriteField("name", d.Name); err != nil {
		return nil, err
	}
	if err := writer.WriteField("isPrivate", strconv.FormatBool(!d.Public)); err != nil {
		return nil, err
	}
	if d.ExternalID != "" {
		if err := writer.WriteField("externalId", d.ExternalID); err != nil {
			return nil, err
		}
	}
	if d.Content != nil {
		part, err := writer.CreateFormFile("content", d.Name)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(d.Content); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return writer, err
	}

	return writer, nil
}
