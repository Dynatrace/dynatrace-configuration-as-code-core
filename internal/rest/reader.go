package rest

import (
	"bytes"
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

func ReusableReader(r io.ReadCloser) io.ReadCloser {
	if r == nil {
		return r
	}
	readBuf := bytes.Buffer{}
	_, _ = readBuf.ReadFrom(r) // nolint: errcheck
	backBuf := bytes.Buffer{}

	return reusableReader{
		io.TeeReader(&readBuf, &backBuf),
		&readBuf,
		&backBuf,
	}

}

func (r reusableReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		r.reset()
	}
	return n, err
}

func (r reusableReader) reset() {
	_, _ = io.Copy(r.readBuf, r.backBuf) // nolint: errcheck
}
