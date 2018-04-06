package util

import (
	"bytes"
	"io"
)

func ReadAll(r io.Reader) (b []byte, err error) {
	buf := BytesBufferPool.Get().(*bytes.Buffer)
	defer BytesBufferPool.Put(buf)

	buf.Reset()

	_, err = buf.ReadFrom(r)
	b = buf.Bytes()

	return
}
