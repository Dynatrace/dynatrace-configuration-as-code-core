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

// formField is a single multipart form field to be written.
type formField struct {
	key, value string
}

// writeDocument serializes the settable fields of meta plus content into a
// multipart body. Read-only Metadata fields are ignored; see Metadata for the
// full list.
func writeDocument(w io.Writer, meta Metadata, content []byte) (contentType, boundary string, err error) {
	writer := multipart.NewWriter(w)
	defer func() {
		if closeErr := writer.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for _, f := range metadataFields(meta) {
		if err := writer.WriteField(f.key, f.value); err != nil {
			return "", "", err
		}
	}

	if content != nil {
		if err := writeContent(writer, meta.Name, content); err != nil {
			return "", "", err
		}
	}

	return writer.FormDataContentType(), writer.Boundary(), nil
}

// metadataFields returns the multipart form fields for the settable metadata.
func metadataFields(meta Metadata) []formField {
	fields := []formField{
		{"type", meta.Type},
		{"name", meta.Name},
		{"isPrivate", strconv.FormatBool(meta.IsPrivate)},
	}
	if meta.ID != "" {
		fields = append(fields, formField{"id", meta.ID})
	}
	if meta.Description != nil {
		fields = append(fields, formField{"description", *meta.Description})
	}
	fields = append(fields, labelFields(meta.Labels)...)
	if meta.IsReshareable != nil {
		fields = append(fields, formField{"isReshareable", strconv.FormatBool(*meta.IsReshareable)})
	}
	return fields
}

// labelFields returns the "labels" form fields. A nil slice yields no fields,
// leaving existing labels untouched; a non-nil but empty slice yields a single
// empty field, which clears them; otherwise one field per label value.
func labelFields(labels []string) []formField {
	if labels == nil {
		return nil
	}
	if len(labels) == 0 {
		return []formField{{"labels", ""}}
	}
	fields := make([]formField, 0, len(labels))
	for _, label := range labels {
		fields = append(fields, formField{"labels", label})
	}
	return fields
}

// writeContent writes the document content as a multipart file part.
func writeContent(writer *multipart.Writer, name string, content []byte) error {
	part, err := writer.CreateFormFile("content", name)
	if err != nil {
		return err
	}
	_, err = part.Write(content)
	return err
}
