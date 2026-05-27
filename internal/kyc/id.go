package kyc

import (
	"crypto/rand"
	"encoding/hex"
)

func newID(prefix string) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return prefix + "_" + hex.EncodeToString(b[:])
}
