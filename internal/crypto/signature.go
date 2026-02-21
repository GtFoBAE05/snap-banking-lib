package crypto

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// ============================================
// KEY PARSING (Required for config validation)
// ============================================

// ParsePrivateKey parses PEM-encoded RSA private key
func ParsePrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	if len(pemData) == 0 {
		return nil, errors.New("empty private key data")
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// Try PKCS8 (most common)
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("not an RSA private key")
	}

	return nil, errors.New("unsupported private key format")
}

// ParsePublicKey parses PEM-encoded RSA public key
func ParsePublicKey(pemData []byte) (*rsa.PublicKey, error) {
	if len(pemData) == 0 {
		return nil, errors.New("empty public key data")
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// Try PKIX (most common)
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("not an RSA public key")
	}

	// Try PKCS1 (legacy)
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, errors.New("unsupported public key format")
}

// ============================================
// RSA SIGNING (Required for Mandiri, BRI, etc)
// ============================================

// SignRSA signs data using RSA private key with SHA-256
// Returns base64-encoded signature
func SignRSA(privateKey *rsa.PrivateKey, data []byte) (string, error) {
	if privateKey == nil {
		return "", errors.New("private key is nil")
	}

	// Hash data
	hash := sha256.Sum256(data)

	// Sign
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("signing failed: %w", err)
	}

	// Return base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

func SignHMAC512(secret string, data []byte) string {
	h := hmac.New(sha512.New, []byte(secret))
	h.Write(data)
	signature := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}

// VerifyRSA verifies RSA signature
func VerifyRSA(publicKey *rsa.PublicKey, data []byte, signatureBase64 string) error {
	if publicKey == nil {
		return errors.New("public key is nil")
	}

	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	// Hash data
	hash := sha256.Sum256(data)

	// Verify
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return errors.New("signature verification failed")
	}

	return nil
}
