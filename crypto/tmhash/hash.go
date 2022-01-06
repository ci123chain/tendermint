package tmhash

import (
	"crypto/sha256"
	"hash"
	"sync"
	"golang.org/x/crypto/sha3"
)

const (
	Size      = sha256.Size
	BlockSize = sha256.BlockSize
)

// New returns a new hash.Hash.
func New() hash.Hash {
	return sha256.New()
}

var keccakPool = sync.Pool{
	// NewLegacyKeccak256 uses non-standard padding
	// and is incompatible with sha3.Sum256
	New: func() interface{} { return sha3.NewLegacyKeccak256() },
}

// Sum returns the non-standard Keccak256 of the bz.
func Sum(bz []byte) []byte {
	sha := keccakPool.Get().(hash.Hash)
	defer func() {
		// better to reset before putting it to the pool
		sha.Reset()
		keccakPool.Put(sha)
	}()
	sha.Reset()
	sha.Write(bz)
	return sha.Sum(nil)
}

//-------------------------------------------------------------

const (
	TruncatedSize = 20
)

type sha256trunc struct {
	sha256 hash.Hash
}

func (h sha256trunc) Write(p []byte) (n int, err error) {
	return h.sha256.Write(p)
}
func (h sha256trunc) Sum(b []byte) []byte {
	shasum := h.sha256.Sum(b)
	return shasum[:TruncatedSize]
}

func (h sha256trunc) Reset() {
	h.sha256.Reset()
}

func (h sha256trunc) Size() int {
	return TruncatedSize
}

func (h sha256trunc) BlockSize() int {
	return h.sha256.BlockSize()
}

// NewTruncated returns a new hash.Hash.
func NewTruncated() hash.Hash {
	return sha256trunc{
		sha256: sha256.New(),
	}
}

// SumTruncated returns the first 20 bytes of SHA256 of the bz.
func SumTruncated(bz []byte) []byte {
	hash := sha256.Sum256(bz)
	return hash[:TruncatedSize]
}
