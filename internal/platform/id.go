package platform

import (
	"crypto/rand"
	"encoding/hex"
)

func ID() string {
	b := make([]byte, 12)
	if _, e := rand.Read(b); e != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
