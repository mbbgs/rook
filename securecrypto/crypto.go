package securecrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	keyLen   = 32      // AES-256
	saltLen  = 16
	nonceLen = 12
	hmacLen  = 32      // SHA-256 output size
	scryptN  = 1 << 15 // scrypt CPU/memory cost
	scryptR  = 8
	scryptP  = 1
)

func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

func HashWithSalt(password, salt string) (string, error) {
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", err
	}

	hash, err := scrypt.Key([]byte(password), saltBytes, 1<<15, 8, 1, 32)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash), nil
}


func VerifyPassword(inputPassword string, hashedPassword string,salt []byte) bool {
	hashedInput, err := HashWithSalt(inputPassword, []byte(salt))
	if err != nil {
		return false
	}
	return hashedInput == hashedPassword
}


func DeriveKey(password []byte, salt []byte) ([]byte, error) {
	return scrypt.Key(password, salt, scryptN, scryptR, scryptP, keyLen)
}



func Encrypt(buffers, key ,salt []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, buffers, nil)

	// Compute HMAC of salt + nonce + ciphertext
	h := hmac.New(sha256.New, key)
	h.Write(salt)
	h.Write(nonce)
	h.Write(ciphertext)
	mac := h.Sum(nil)

	full := append(salt, append(nonce, append(ciphertext, mac...)...)...)
	return full, nil
}


func Decrypt(payload, key []byte) ([]byte, error) {
	if len(payload) < saltLen+nonceLen+hmacLen {
		return nil, errors.New("payload too short")
	}

	salt := payload[:saltLen]
	nonce := payload[saltLen : saltLen+nonceLen]
	macStart := len(payload) - hmacLen
	ciphertext := payload[saltLen+nonceLen : macStart]
	expectedMAC := payload[macStart:]


	// Check HMAC
	h := hmac.New(sha256.New, key)
	h.Write(salt)
	h.Write(nonce)
	h.Write(ciphertext)
	actualMAC := h.Sum(nil)

	if !hmac.Equal(actualMAC, expectedMAC) {
		return nil, errors.New("HMAC mismatch: tampered or wrong password")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	buffers, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed or integrity check failed")
	}

	return buffers, nil
}
