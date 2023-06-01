package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

const (
	PEM_LOCATION        = "./auth/bin/rsapems/"
	PUBLIC_KEY_LOCATION = "./auth/bin/rsapems/public.pem"
	PATH_TO_BINS        = "./auth/bin/"
	ERR                 = "The system cannot find the file specified"
)

func GenerateKeyPair(withForce bool) error {
	if _, err := os.Stat(PUBLIC_KEY_LOCATION); !errors.Is(err, os.ErrNotExist) && !withForce {
		return nil
	}
	if withForce {
		err := os.Remove(PEM_LOCATION + "private.pem")
		if err != nil {
			return err
		}
		err = os.Remove(PUBLIC_KEY_LOCATION)
		if err != nil {
			return err
		}
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privateKeyFile, err := os.Create(PEM_LOCATION + "private.pem")
	if err != nil {
		return err
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	err = pem.Encode(privateKeyFile, privateKeyPEM)
	if err != nil {
		return err
	}
	privateKeyFile.Close()

	// Save public key to file
	publicKeyFile, err := os.Create(PEM_LOCATION + "public.pem")
	if err != nil {
		return err
	}

	publicKey := privateKey.PublicKey
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}

	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	err = pem.Encode(publicKeyFile, publicKeyPEM)
	if err != nil {
		return err
	}
	publicKeyFile.Close()

	return nil
}

func EncryptData(data []byte, publicKeyFile string) (string, error) {
	publicKeyBytes, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return "", errors.New("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey.(*rsa.PublicKey), data)
	if err != nil {
		return "", err
	}

	return string(encryptedData), nil
}

func DecryptData(encryptedData []byte) ([]byte, error) {
	privateKeyBytes, err := os.ReadFile(PEM_LOCATION)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
