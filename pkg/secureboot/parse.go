package secureboot

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

func ParseHash(input string) ([32]byte, error) {
	var hash [32]byte

	s := strings.TrimSpace(input)
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	if len(s) != 64 {
		return hash, fmt.Errorf("expected 64 hex characters, got %d", len(s))
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return hash, err
	}
	copy(hash[:], decoded)
	return hash, nil
}

func ValidateKeySize(keySize int) error {
	if keySize != 2048 && keySize != 4096 {
		return errors.New("must be 2048 or 4096")
	}
	return nil
}
