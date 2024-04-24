package cert

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

type selfSigner struct {
	organisation string
	hosts        []string

	serial *big.Int

	caCert     []byte
	caTemplate *x509.Certificate
	caSigner   crypto.Signer
}

func NewSelfSigner(organisation string, hosts []string) (Source, error) {
	if len(hosts) == 0 {
		return nil, ErrUnspecifiedHosts
	}

	serial, err := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(128), nil))
	if err != nil {
		return nil, err
	}

	return &selfSigner{
		organisation: organisation,
		hosts:        hosts,

		serial: serial,
	}, nil
}

func (s *selfSigner) NewBundle() (*Bundle, error) {
	if s.caCert == nil {
		if err := s.regenerateCA(); err != nil {
			return nil, err
		}
	}

	cert, err := s.generateCert()
	if err != nil {
		return nil, err
	}

	return &Bundle{
		CA:   bytes.Clone(s.caCert),
		Pair: cert,
	}, nil
}

func (s *selfSigner) newEcPrivateKey() (string, crypto.Signer, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrFailedToGeneratePrivateKey, err)
	}

	bts, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrFailedToGeneratePrivateKey, err)
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: bts})
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrFailedToGeneratePrivateKey, err)
	}

	return buf.String(), key, nil
}

func (s *selfSigner) regenerateCA() error {
	recently := time.Now().AddDate(0, 0, -1).Round(time.Hour)

	_, sgn, err := s.newEcPrivateKey()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToRegenerateCA, err)
	}

	tpl := &x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  true,
		SerialNumber:          s.serial.Add(s.serial, big.NewInt(1)),
		Subject:               pkix.Name{Organization: []string{s.organisation}},

		KeyUsage:    x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		NotBefore: recently,
		NotAfter:  recently.AddDate(1, 0, 0),
	}

	bts, err := x509.CreateCertificate(rand.Reader, tpl, tpl, sgn.Public(), sgn)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToRegenerateCA, err)
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: bts})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToRegenerateCA, err)
	}

	s.caSigner = sgn
	s.caCert = buf.Bytes()
	s.caTemplate = tpl

	return nil
}

func (s *selfSigner) generateCert() (*tls.Certificate, error) {
	recently := time.Now().AddDate(0, 0, -1).Round(time.Hour)

	key, sgn, err := s.newEcPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGenerateCert, err)
	}

	cert := &x509.Certificate{
		BasicConstraintsValid: true,
		SerialNumber:          s.serial.Add(s.serial, big.NewInt(1)),
		Subject:               pkix.Name{CommonName: s.hosts[0], Organization: []string{s.organisation}},

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		NotBefore: recently,
		NotAfter:  recently.AddDate(1, 0, 0),
	}

	for _, h := range s.hosts {
		if ip := net.ParseIP(h); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, h)
		}
	}

	bts, err := x509.CreateCertificate(rand.Reader, cert, s.caTemplate, sgn.Public(), s.caSigner)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGenerateCert, err)
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: bts})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGenerateCert, err)
	}

	pair, err := tls.X509KeyPair(buf.Bytes(), []byte(key))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGenerateCert, err)
	}

	return &pair, nil
}
