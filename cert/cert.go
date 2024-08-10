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
	errFailedToGenerateCert       = errors.New("failed to generate certificate")
	errFailedToGeneratePrivateKey = errors.New("failed to generate new private key")
	errFailedToRegenerateCA       = errors.New("failed to (re-)generate ca")
	errUnspecifiedHosts           = errors.New("no hosts specified for the certificate")
)
