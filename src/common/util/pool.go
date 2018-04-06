package util

import (
	"bytes"
	"sync"
)

var BytesBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
