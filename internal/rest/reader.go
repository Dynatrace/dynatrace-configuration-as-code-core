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

package rest

import (
	"bytes"
	"fmt"
	"io"
)

// reusableReader is a reader that can be used multiple times to from an io.ReadCloser
type reusableReader struct {
	io.Reader
	readBuf *bytes.Buffer
	backBuf *bytes.Buffer
}

func (r reusableReader) Close() error {
	return nil
}

func ReusableReader(r io.ReadCloser) (io.ReadCloser, error) {
	if r == nil {
		return r, nil
	}
	readBuf := bytes.Buffer{}
	if _, err := readBuf.ReadFrom(r); err != nil {
		return nil, err
	}
	backBuf := bytes.Buffer{}
	return reusableReader{
		io.TeeReader(&readBuf, &backBuf),
		&readBuf,
		&backBuf,
	}, nil

}

func (r reusableReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		if err := r.reset(); err != nil {
			return n, fmt.Errorf("failed to reset reader after reaching EOF: %w", err)
		}
	}
	return n, err
}

func (r reusableReader) reset() error {
	if _, err := io.Copy(r.readBuf, r.backBuf); err != nil {
		return err
	}
	return nil
}
