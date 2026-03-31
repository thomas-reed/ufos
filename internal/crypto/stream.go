package crypto

import (
	"crypto/cipher"
	"io"
)

// Wraps an io.Writer and performs AES-CTR encryption on the fly
type EncryptingWriter struct {
	W      io.Writer
	Stream cipher.Stream
}

func (e *EncryptingWriter) Write(payload []byte) (n int, err error) {
	e.Stream.XORKeyStream(payload, payload)
	return e.W.Write(payload)
}

// Wraps an io.Reader to decrypt on the fly
type DecryptingReader struct {
	R      io.Reader
	Stream cipher.Stream
}

func (d *DecryptingReader) Read(payload []byte) (n int, err error) {
	n, err = d.R.Read(payload)
	if n > 0 {
		d.Stream.XORKeyStream(payload[:n], payload[:n])
	}
	return n, err
}
