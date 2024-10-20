package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"golang.org/x/crypto/scrypt"
)

func deriveKey(password, salt []byte, keyLen int) ([]byte, error) {
	// scrypt parameters: N=32768, r=8, p=1
	return scrypt.Key(password, salt, 32768, 8, 1, keyLen)
}
func Encrypt(plaintext, password []byte) ([]byte, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Derive a key from the password and salt
	key, err := deriveKey(password, salt, 32) // 32 bytes for AES-256
	if err != nil {
		return nil, err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create AES-GCM
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a nonce (GCM standard nonce size is 12 bytes)
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the plaintext
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Return salt + nonce + ciphertext as base64
	finalData := append(salt, nonce...)
	finalData = append(finalData, ciphertext...)
	return finalData, nil
}

func Decrypt(ciphertext []byte, password []byte) ([]byte, error) {

	// Extract salt (first 16 bytes)
	salt := ciphertext[:16]

	// Derive the key from the password and salt
	key, err := deriveKey(password, salt, 32) // 32 bytes for AES-256
	if err != nil {
		return nil, err
	}

	// Extract nonce (next 12 bytes)
	nonce := ciphertext[16 : 16+12]

	// Extract the actual ciphertext
	actualCiphertext := ciphertext[16+12:]

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create AES-GCM
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	plaintext, err := aesgcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, err
	}

	return (plaintext), nil
}
