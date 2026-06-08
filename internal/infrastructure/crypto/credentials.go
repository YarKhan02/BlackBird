package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

func GenerateClientID(name string) (string, error) {
	slug := slugify(name)
	if slug == "" {
		slug = "app"
	}
	suffix, err := randomBase32(10) // 50 bits
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("shk_%s_%s", slug, suffix), nil
}

func GenerateClientSecret() (string, error) {
	b := make([]byte, 32) // 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var out []rune
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
			lastDash = false
		case !lastDash:
			out = append(out, '-')
			lastDash = true
		}
	}
	return strings.Trim(string(out), "-")
}

func randomBase32(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}
