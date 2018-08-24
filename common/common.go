package common

import (
	"sync"
	"bytes"
)

var BytesBufferPool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

