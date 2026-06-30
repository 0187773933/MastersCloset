// Package encryption ports the crypto helpers v1 relied on, unchanged in
// behavior so existing BoltDB ciphertext stays readable. Test/demo helpers
// from v1 were dropped; only the primitives the app actually uses remain.
package encryption

import (
	random "crypto/rand"
	base64 "encoding/base64"
	hex "encoding/hex"
	"io"

	bcrypt "golang.org/x/crypto/bcrypt"
	chacha "golang.org/x/crypto/chacha20poly1305"
	secretbox "golang.org/x/crypto/nacl/secretbox"
)

// GenerateRandomString returns byte_length random bytes, hex-encoded.
func GenerateRandomString(byteLength int) string {
	b := make([]byte, byteLength)
	random.Read(b)
	return hex.EncodeToString(b)
}

// SecretBoxGenerateKey derives a 32-byte key from a password via bcrypt.
func SecretBoxGenerateKey(password string) (key [32]byte) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost+3)
	copy(key[:], hashed[:32])
	return
}

// SecretBoxEncrypt seals plain_text with a hex-encoded 32-byte key.
func SecretBoxEncrypt(key string, plainText string) string {
	keyHex, _ := hex.DecodeString(key)
	var keyBytes [32]byte
	copy(keyBytes[:], keyHex)
	var nonce [24]byte
	io.ReadFull(random.Reader, nonce[:])
	encrypted := secretbox.Seal(nonce[:], []byte(plainText), &nonce, &keyBytes)
	return base64.StdEncoding.EncodeToString(encrypted)
}

// SecretBoxDecrypt reverses SecretBoxEncrypt. Returns "" on failure.
func SecretBoxDecrypt(key string, encrypted string) string {
	keyHex, _ := hex.DecodeString(key)
	var keyBytes [32]byte
	copy(keyBytes[:], keyHex)
	encryptedBytes, _ := base64.StdEncoding.DecodeString(encrypted)
	if len(encryptedBytes) < 24 {
		return ""
	}
	var nonce [24]byte
	copy(nonce[:], encryptedBytes[0:24])
	decrypted, ok := secretbox.Open(nil, encryptedBytes[24:], &nonce, &keyBytes)
	if !ok {
		return ""
	}
	return string(decrypted)
}

// ChaChaEncryptBytes seals plaintext bytes with a hex-encoded 32-byte key,
// prefixing the nonce. Used for at-rest user records in BoltDB.
func ChaChaEncryptBytes(key string, plainText []byte) []byte {
	keyHex, _ := hex.DecodeString(key)
	var keyBytes [32]byte
	copy(keyBytes[:], keyHex)
	aead, _ := chacha.New(keyBytes[:])
	nonce := make([]byte, aead.NonceSize())
	io.ReadFull(random.Reader, nonce[:])
	encrypted := aead.Seal(nil, nonce, plainText, nil)
	return append(nonce, encrypted...)
}

// ChaChaDecryptBytes reverses ChaChaEncryptBytes. Returns nil on failure.
func ChaChaDecryptBytes(key string, encrypted []byte) []byte {
	keyHex, _ := hex.DecodeString(key)
	var keyBytes [32]byte
	copy(keyBytes[:], keyHex)
	aead, _ := chacha.New(keyBytes[:])
	ns := aead.NonceSize()
	if len(encrypted) < ns {
		return nil
	}
	nonce := make([]byte, ns)
	copy(nonce[:], encrypted[0:ns])
	decrypted, err := aead.Open(nil, nonce, encrypted[ns:], nil)
	if err != nil {
		return nil
	}
	return decrypted
}
