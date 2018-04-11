package util

import (
	"bytes"
	"io"
)

func ReadAll(r io.Reader) (b []byte, err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()

	buf := BytesBufferPool.Get().(*bytes.Buffer)
	defer BytesBufferPool.Put(buf)

	buf.Reset()

	_, err = buf.ReadFrom(r)
	b = buf.Bytes()

	return
}
