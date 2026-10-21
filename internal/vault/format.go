package vault

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

// File layout (integers are big-endian):
//
//	header  = magic "DTVLT" | version u8 | memory KiB u32 | iterations u32 | threads u8 | salt [16]
//	wrapped = nonce [24] | XChaCha20-Poly1305(KEK, DEK, aad = header)
//	payload = nonce [24] | XChaCha20-Poly1305(DEK, JSON, aad = header)
//
// The KEK is derived from the keyword and the password with Argon2id. The DEK
// is random and encrypts the data, so changing credentials only rewraps the DEK.
const (
	magic       = "DTVLT"
	version     = 1
	saltSize    = 16
	keySize     = chacha20poly1305.KeySize
	nonceSize   = chacha20poly1305.NonceSizeX
	headerSize  = len(magic) + 1 + 4 + 4 + 1 + saltSize
	wrappedSize = nonceSize + keySize + chacha20poly1305.Overhead
)

// KDFParams are the Argon2id parameters stored in the vault header.
type KDFParams struct {
	MemoryKiB  uint32
	Iterations uint32
	Threads    uint8
}

// DefaultKDF is used for new vaults.
var DefaultKDF = KDFParams{MemoryKiB: 64 * 1024, Iterations: 3, Threads: 4}

// Bounds for parameters read from a file, so a crafted header cannot make
// Argon2 allocate an absurd amount of memory or run forever.
const (
	maxMemoryKiB  = 4 * 1024 * 1024
	maxIterations = 64
)

func (p KDFParams) validate() error {
	if p.MemoryKiB < 8*uint32(p.Threads) || p.MemoryKiB > maxMemoryKiB ||
		p.Iterations < 1 || p.Iterations > maxIterations || p.Threads < 1 {
		return fmt.Errorf("%w: invalid kdf parameters", ErrCorrupt)
	}
	return nil
}

func encodeHeader(p KDFParams, salt []byte) []byte {
	h := make([]byte, 0, headerSize)
	h = append(h, magic...)
	h = append(h, version)
	h = binary.BigEndian.AppendUint32(h, p.MemoryKiB)
	h = binary.BigEndian.AppendUint32(h, p.Iterations)
	h = append(h, p.Threads)
	return append(h, salt...)
}

func decodeHeader(h []byte) (KDFParams, []byte, error) {
	if len(h) != headerSize || string(h[:len(magic)]) != magic {
		return KDFParams{}, nil, ErrCorrupt
	}
	o := len(magic)
	if h[o] != version {
		return KDFParams{}, nil, fmt.Errorf("unsupported vault version %d", h[o])
	}
	o++
	p := KDFParams{
		MemoryKiB:  binary.BigEndian.Uint32(h[o:]),
		Iterations: binary.BigEndian.Uint32(h[o+4:]),
		Threads:    h[o+8],
	}
	if err := p.validate(); err != nil {
		return KDFParams{}, nil, err
	}
	return p, h[o+9:], nil
}

// deriveKEK length-prefixes both secrets so ("ab", "c") and ("a", "bc") never
// produce the same key.
func deriveKEK(keyword, password, salt []byte, p KDFParams) []byte {
	in := make([]byte, 0, 8+len(keyword)+len(password))
	in = binary.BigEndian.AppendUint32(in, uint32(len(keyword)))
	in = append(in, keyword...)
	in = binary.BigEndian.AppendUint32(in, uint32(len(password)))
	in = append(in, password...)
	defer wipe(in)
	return argon2.IDKey(in, salt, p.Iterations, p.MemoryKiB, p.Threads, keySize)
}

func seal(key, plaintext, aad []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	nonce := randomBytes(nonceSize)
	return aead.Seal(nonce, nonce, plaintext, aad), nil
}

func open(key, box, aad []byte) ([]byte, error) {
	if len(box) < nonceSize+chacha20poly1305.Overhead {
		return nil, ErrCorrupt
	}
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, box[:nonceSize], box[nonceSize:], aad)
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	// crypto/rand.Read never returns an error on supported platforms.
	_, _ = rand.Read(b)
	return b
}

func wipe(b []byte) {
	clear(b)
}
