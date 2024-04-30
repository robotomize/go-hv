package mergetool

import (
	"fmt"

	"github.com/robotomize/go-bf/bf/bfbits"
)

type keyFunc func(ts int64, command string) []byte

type bloomFilter interface {
	Add(item []byte) error
	Contains(item []byte) (bool, error)
}

func newStorage(bloomFilter bloomFilter, keyFn keyFunc) *storage {
	return &storage{bf: bloomFilter, keyFn: keyFn}
}

type storage struct {
	bf    bfbits.BloomFilter
	keyFn keyFunc
}

func (s *storage) Add(ts int64, command string) error {
	key := s.keyFn(ts, command)
	if err := s.bf.Add(key); err != nil {
		return fmt.Errorf("bloom filter Add: %w", err)
	}
	return nil
}

func (s *storage) Contains(ts int64, command string) (bool, error) {
	key := s.keyFn(ts, command)
	ok, err := s.bf.Contains(key)
	if err != nil {
		return false, fmt.Errorf("bloom filter Add: %w", err)
	}
	return ok, nil
}
