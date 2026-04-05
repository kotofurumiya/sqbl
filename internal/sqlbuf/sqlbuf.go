// Package sqlbuf provides a pooled bytes.Buffer for renderSQL methods.
// Using a pool avoids re-allocating the buffer on every ToSql call.
//
// bytes.Buffer.Reset() retains the underlying []byte (unlike strings.Builder.Reset()
// which sets buf to nil), so pool reuse actually eliminates the Grow allocation
// on subsequent calls.
package sqlbuf

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// GetStringBuffer returns a reset Buffer from the pool.
func GetStringBuffer() *bytes.Buffer {
	b := pool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

// maxPoolBufCap is the maximum buffer capacity retained in the pool.
// Buffers that grew beyond this threshold are discarded so that one large query
// does not permanently occupy pool slots with oversized backing arrays.
const maxPoolBufCap = 1 * 1024

// PutStringBuffer returns b to the pool. b must not be used after this call.
// Oversized buffers are discarded rather than pooled.
func PutStringBuffer(b *bytes.Buffer) {
	if b.Cap() > maxPoolBufCap {
		return
	}
	pool.Put(b)
}
