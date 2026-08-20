package platform

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func HashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func HashReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, e := io.Copy(h, r); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func HashReaderContext(ctx context.Context, r io.Reader) (string, error) {
	return HashReader(r)
}
