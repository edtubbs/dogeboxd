package secureboot

import "testing"

func TestParseHash(t *testing.T) {
	hash, err := ParseHash("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash[0] != 0x00 || hash[1] != 0x11 || hash[31] != 0xff {
		t.Fatalf("unexpected hash bytes: %x", hash)
	}
}

func TestParseHashRejectsInvalidLength(t *testing.T) {
	_, err := ParseHash("abcd")
	if err == nil {
		t.Fatalf("expected error for invalid length hash")
	}
}

func TestValidateKeySize(t *testing.T) {
	if err := ValidateKeySize(2048); err != nil {
		t.Fatalf("unexpected error for 2048: %v", err)
	}

	if err := ValidateKeySize(4096); err != nil {
		t.Fatalf("unexpected error for 4096: %v", err)
	}

	if err := ValidateKeySize(3072); err == nil {
		t.Fatal("expected error for unsupported key size")
	}
}
