package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"

	"github.com/flawiddsouza/git-share/internal/wordlist"
)

const (
	// CodeIDLength is the length of the random code ID used for server lookups.
	CodeIDLength = 10
	// PassphraseWords is the number of diceware words in a passphrase.
	PassphraseWords = 4
	// PassphraseSep is the separator between words in a passphrase.
	PassphraseSep = "-"
	// CodeSep separates the code ID from the passphrase in a combined code.
	CodeSep = "-"
	// hkdfSalt is a fixed salt for HKDF key derivation.
	hkdfSalt = "git-share-v1"
	// hkdfInfo is the context info for HKDF key derivation.
	hkdfInfo = "encryption-key"
)

// base62 charset for generating code IDs.
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GenerateCode creates a combined code: <codeId>-<word1>-<word2>-<word3>-<word4>.
// The codeId is a random base62 string used for server lookup.
// The passphrase is used for key derivation / encryption.
func GenerateCode() (code string, codeID string, passphrase string, err error) {
	codeID, err = generateCodeID()
	if err != nil {
		return "", "", "", fmt.Errorf("generating code ID: %w", err)
	}

	passphrase, err = wordlist.Pick(PassphraseWords, PassphraseSep)
	if err != nil {
		return "", "", "", fmt.Errorf("generating passphrase: %w", err)
	}

	code = codeID + CodeSep + passphrase
	return code, codeID, passphrase, nil
}

// ParseCode splits a combined code into codeID and passphrase.
// Format: <codeId>-<word1>-<word2>-<word3>-<word4>
func ParseCode(code string) (codeID string, passphrase string, err error) {
	parts := strings.SplitN(code, CodeSep, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("invalid code format: expected <codeId>-<word1>-<word2>-<word3>-<word4>")
	}

	// Validate that passphrase has the expected number of words
	words := strings.Split(parts[1], PassphraseSep)
	if len(words) != PassphraseWords {
		return "", "", fmt.Errorf("invalid code format: passphrase should have %d words, got %d", PassphraseWords, len(words))
	}

	return parts[0], parts[1], nil
}

// DeriveKey derives a 256-bit encryption key from a passphrase using HKDF-SHA256.
func DeriveKey(passphrase string) ([]byte, error) {
	hkdfReader := hkdf.New(sha256.New, []byte(passphrase), []byte(hkdfSalt), []byte(hkdfInfo))
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, fmt.Errorf("deriving key: %w", err)
	}
	return key, nil
}

// Encrypt encrypts plaintext using XChaCha20-Poly1305.
// Returns: nonce || ciphertext (includes auth tag).
func Encrypt(plaintext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	// Seal appends the ciphertext and tag to the nonce
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext produced by Encrypt using XChaCha20-Poly1305.
// Input format: nonce || ciphertext (includes auth tag).
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	encrypted := ciphertext[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong passphrase?): %w", err)
	}

	return plaintext, nil
}

// generateCodeID creates a random base62 string of CodeIDLength.
func generateCodeID() (string, error) {
	max := big.NewInt(int64(len(base62Chars)))
	b := make([]byte, CodeIDLength)
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = base62Chars[idx.Int64()]
	}
	return string(b), nil
}
