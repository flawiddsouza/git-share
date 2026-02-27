package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	code, codeID, passphrase, err := GenerateCode()
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}
	if len(codeID) != CodeIDLength {
		t.Errorf("expected codeID length %d, got %d", CodeIDLength, len(codeID))
	}
	if passphrase == "" {
		t.Error("expected non-empty passphrase")
	}
	if code == "" {
		t.Error("expected non-empty code")
	}
	t.Logf("Generated code: %s", code)
}

func TestParseCode(t *testing.T) {
	code, codeID, passphrase, err := GenerateCode()
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}

	parsedID, parsedPass, err := ParseCode(code)
	if err != nil {
		t.Fatalf("ParseCode(%q) error: %v", code, err)
	}
	if parsedID != codeID {
		t.Errorf("codeID mismatch: got %q, want %q", parsedID, codeID)
	}
	if parsedPass != passphrase {
		t.Errorf("passphrase mismatch: got %q, want %q", parsedPass, passphrase)
	}
}

func TestParseCodeInvalid(t *testing.T) {
	cases := []string{
		"",
		"nodelimiter",
		"-",
		"id-",
		"-passphrase",
		"id-oneword",
	}
	for _, c := range cases {
		_, _, err := ParseCode(c)
		if err == nil {
			t.Errorf("ParseCode(%q) expected error, got nil", c)
		}
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := []byte("this is a git patch\n--- a/file.go\n+++ b/file.go\n")

	key, err := DeriveKey("alpha-bravo-charlie-delta")
	if err != nil {
		t.Fatalf("DeriveKey() error: %v", err)
	}

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Error("ciphertext should not equal plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted text mismatch:\ngot:  %q\nwant: %q", decrypted, plaintext)
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	plaintext := []byte("secret patch data")

	key1, _ := DeriveKey("correct-horse-battery-staple")
	key2, _ := DeriveKey("wrong-passphrase-here-now")

	ciphertext, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	_, err = Decrypt(ciphertext, key2)
	if err == nil {
		t.Error("expected decryption to fail with wrong key, but it succeeded")
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	key1, _ := DeriveKey("same-passphrase-every-time")
	key2, _ := DeriveKey("same-passphrase-every-time")

	if !bytes.Equal(key1, key2) {
		t.Error("same passphrase should produce the same key")
	}
}

func TestDeriveKeyDifferentPassphrases(t *testing.T) {
	key1, _ := DeriveKey("alpha-bravo-charlie-delta")
	key2, _ := DeriveKey("echo-foxtrot-golf-hotel")

	if bytes.Equal(key1, key2) {
		t.Error("different passphrases should produce different keys")
	}
}
