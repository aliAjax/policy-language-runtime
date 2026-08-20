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
	h := sha256.New()
	buffer := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		n, err := r.Read(buffer)
		if n > 0 {
			if _, writeErr := h.Write(buffer[:n]); writeErr != nil {
				return "", writeErr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
