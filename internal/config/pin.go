package config

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// MinPIN is the shortest PIN accepted, in characters.
const MinPIN = 4

// The PIN only gates starting dtvlt; it never touches the vault key. It is
// stored as an Argon2id hash in PHC string format.
const (
	pinMemoryKiB  = 32 * 1024
	pinIterations = 2
	pinThreads    = 2
	pinKeyLen     = 32
)

func HashPIN(pin []byte) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	key := argon2.IDKey(pin, salt, pinIterations, pinMemoryKiB, pinThreads, pinKeyLen)
	enc := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, pinMemoryKiB, pinIterations, pinThreads, enc.EncodeToString(salt), enc.EncodeToString(key))
}

func VerifyPIN(encoded string, pin []byte) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version int
	var m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil || m > 1<<20 || t > 16 || p == 0 {
		return false
	}
	enc := base64.RawStdEncoding
	salt, err := enc.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := enc.DecodeString(parts[5])
	if err != nil {
		return false
	}
	got := argon2.IDKey(pin, salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}
