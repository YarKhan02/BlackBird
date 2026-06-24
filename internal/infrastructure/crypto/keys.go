package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)


func LoadRSAPrivateKeyFromPEM(pemData string) (*rsa.PrivateKey, error) {
	data := strings.TrimSpace(pemData)
	if data == "" {
		return nil, fmt.Errorf("RSA_PRIVATE_KEY_PEM is empty")
	}

	return parseRSAPrivateKey([]byte(data), "RSA_PRIVATE_KEY_PEM")
}

func parseRSAPrivateKey(data []byte, source string) (*rsa.PrivateKey, error) {

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM data in %s", source)
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	return rsaKey, nil
}
