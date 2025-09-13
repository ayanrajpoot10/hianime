package scraper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// CryptoJS compatible AES decryption
func aesDecrypt(encrypted, passphrase string) (string, error) {
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check for "Salted__" prefix (CryptoJS format)
	if len(data) < 16 || string(data[:8]) != "Salted__" {
		return "", fmt.Errorf("invalid encrypted data format")
	}

	// Extract salt
	salt := data[8:16]
	ciphertext := data[16:]

	// Derive key and IV using PBKDF2 (CryptoJS compatible)
	passBytes := []byte(passphrase)
	keyIV := pbkdf2.Key(passBytes, salt, 1000, 48, sha1.New) // 32 bytes key + 16 bytes IV
	key := keyIV[:32]
	iv := keyIV[32:48]

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt using CBC mode
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > len(plaintext) || padding > aes.BlockSize {
		return "", fmt.Errorf("invalid padding")
	}

	for i := len(plaintext) - padding; i < len(plaintext); i++ {
		if plaintext[i] != byte(padding) {
			return "", fmt.Errorf("invalid padding")
		}
	}

	result := plaintext[:len(plaintext)-padding]
	return string(result), nil
}

// Simple AES decryption for cases where data is already properly formatted
func simpleAESDecrypt(encrypted, key string) (string, error) {
	// Remove any whitespace
	key = strings.TrimSpace(key)

	// Try direct AES decryption first
	if result, err := aesDecrypt(encrypted, key); err == nil {
		return result, nil
	}

	// Fallback: try with different key formats if needed
	return "", fmt.Errorf("failed to decrypt with provided key")
}
