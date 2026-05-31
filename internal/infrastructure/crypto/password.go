package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id — winner of Password Hashing Competition, better than bcrypt
const (
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	var mem, itr uint32
	var par uint8
	fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &itr, &par)

	salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
	decodedHash, _ := base64.RawStdEncoding.DecodeString(parts[5])

	compHash := argon2.IDKey([]byte(password), salt, itr, mem, par, uint32(len(decodedHash)))

	return subtle.ConstantTimeCompare(compHash, decodedHash) == 1
}
