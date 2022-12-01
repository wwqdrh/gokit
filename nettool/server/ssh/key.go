package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/wwqdrh/logger"
)

// SSHGenerator ssh key pair generator
type SSHGenerator struct {
	PrivateKey, PublicKey []byte
	PrivateKeyPath        string
}

// NewSSHGenerator create ssh generator
func NewSSHGenerator(privateKey string, publicKey string, privateKeyPath string) *SSHGenerator {
	return &SSHGenerator{
		PrivateKey:     []byte(privateKey),
		PublicKey:      []byte(publicKey),
		PrivateKeyPath: privateKeyPath,
	}
}

// Generate generate SSHGenerator
func Generate(privateKeyPath string) (*SSHGenerator, error) {
	privateKey, err := generatePrivateKey(SshBitSize)
	if err != nil {
		return nil, err
	}

	publicKeyBytes, err := encodePublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	sshKey := &SSHGenerator{
		PrivateKey:     privateKeyBytes,
		PrivateKeyPath: privateKeyPath,
		PublicKey:      publicKeyBytes,
	}
	_ = os.Remove(sshKey.PrivateKeyPath)
	err = WritePrivateKey(sshKey.PrivateKeyPath, sshKey.PrivateKey)
	return sshKey, err
}

// PrivateKeyPath ...
func PrivateKeyPath(name string) string {
	return fmt.Sprintf("%s/%s%s", NettoolKeyDir, name, PostfixRsaKey)
}

// CleanRsaKeys ...
func CleanRsaKeys() {
	files, _ := ioutil.ReadDir(NettoolKeyDir)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), PostfixRsaKey) {
			rsaKey := fmt.Sprintf("%s/%s", NettoolKeyDir, f.Name())
			err := os.Remove(rsaKey)
			if err != nil {
				logger.DefaultLogger.Debug("Failed to remove rsa key file: " + rsaKey)
			} else {
				logger.DefaultLogger.Debugx("Unused rsa key %s removed", nil, f.Name())
			}
		}
	}
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	logger.DefaultLogger.Debug("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// encodePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// return format "ssh-rsa ..."
func encodePublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)
	logger.DefaultLogger.Debug("Public key generated")
	return pubKeyBytes, nil
}

// WritePrivateKey write ssh private key to privateKeyPath
func WritePrivateKey(privateKeyPath string, data []byte) error {
	dir := filepath.Dir(privateKeyPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0700); err != nil {
			logger.DefaultLogger.Error("Can't create dir " + dir)
			return err
		}
	}
	if err := ioutil.WriteFile(privateKeyPath, data, 0400); err != nil {
		logger.DefaultLogger.Errorx("Write ssh private key to %s failed", nil, privateKeyPath)
		return err
	}
	return nil
}
