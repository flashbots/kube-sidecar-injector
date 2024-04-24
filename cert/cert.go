package cert

import (
	"crypto/tls"
	"errors"
)

type Bundle struct {
	CA   []byte
	Pair *tls.Certificate
}

type Source interface {
	NewBundle() (*Bundle, error)
}

var (
	ErrFailedToGenerateCert       = errors.New("failed to generate certificate")
	ErrFailedToGeneratePrivateKey = errors.New("failed to generate new private key")
	ErrFailedToRegenerateCA       = errors.New("failed to (re-)generate ca")
	ErrUnspecifiedHosts           = errors.New("no hosts specified for the certificate")
)
