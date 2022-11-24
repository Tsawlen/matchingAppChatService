package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

var publicKey *rsa.PublicKey

func GetPublicToken() (*rsa.PublicKey, error) {
	if publicKey == nil {
		key, err := getPublicTokenFromEnvironment()
		if err != nil {
			return nil, err
		}
		return key, nil
	}
	return publicKey, nil
}

func getPublicTokenFromEnvironment() (*rsa.PublicKey, error) {
	envToken := os.Getenv("PUBLIC_SECRET")

	block, _ := pem.Decode([]byte(envToken))
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("Could not extract public key!")
	}
	var keyReturn = key.(*rsa.PublicKey)
	return keyReturn, nil
}
