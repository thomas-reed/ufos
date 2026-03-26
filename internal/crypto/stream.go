package crypto

import (
	"crypto/cipher"
	"io"
)

// EncryptingWriter wraps an io.Writer and performs AES-CTR encryption on the fly
type EncryptingWriter struct {
	W      io.Writer
	Stream cipher.Stream
}

func (e *EncryptingWriter) Write(p []byte) (n int, err error) {
	e.Stream.XORKeyStream(p, p)
	return e.W.Write(p)
}