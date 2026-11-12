package sso

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// The service provider's signing key (V4-03). Each workspace's SAML service
// provider has its own: an RSA-2048 key with a self-signed certificate, which
// is what SAML metadata exchanges (the IdP trusts the certificate it was given,
// not a CA). It signs AuthnRequests when the connection asks and every logout
// message. A rotation stages a next key, published in the metadata beside the
// current one, and activating it makes it the one that signs.

// SPKeyLife is how long a service-provider certificate is valid.
const SPKeyLife = 5 * 365 * 24 * time.Hour

// SPKeyPair is a service provider's key and certificate.
type SPKeyPair struct {
	Key  *rsa.PrivateKey
	Cert *x509.Certificate
}

// NewSPKeyPair makes a key and a certificate naming the service provider.
func NewSPKeyPair(entityID string) (*SPKeyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		return nil, err
	}
	cn := entityID
	if len(cn) > 64 {
		cn = cn[len(cn)-64:]
	}
	now := time.Now().UTC()
	tpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(SPKeyLife),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}
	return &SPKeyPair{Key: key, Cert: cert}, nil
}

// KeyPEM is the private key, PKCS#8 PEM (seal it before storing it).
func (k *SPKeyPair) KeyPEM() string {
	der, _ := x509.MarshalPKCS8PrivateKey(k.Key)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

// CertB64 is the certificate, base64 DER as metadata carries it.
func (k *SPKeyPair) CertB64() string { return base64.StdEncoding.EncodeToString(k.Cert.Raw) }

// CertPEM is the certificate as PEM, for an IdP admin who uploads a file.
func (k *SPKeyPair) CertPEM() string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: k.Cert.Raw}))
}

// Fingerprint is the certificate's SHA-256, colon-separated hex (as IdP consoles show it).
func (k *SPKeyPair) Fingerprint() string { return CertFingerprint(k.Cert) }

// CertFingerprint is a certificate's SHA-256, colon-separated upper-case hex.
func CertFingerprint(c *x509.Certificate) string {
	if c == nil {
		return ""
	}
	sum := sha256.Sum256(c.Raw)
	parts := make([]string, len(sum))
	for i, b := range sum {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":")
}

// ParseSPKeyPair reads a key (PKCS#8 or PKCS#1 PEM) and its certificate (base64 DER) back.
func ParseSPKeyPair(keyPEM, certB64 string) (*SPKeyPair, error) {
	blk, _ := pem.Decode([]byte(keyPEM))
	if blk == nil {
		return nil, errors.New("sso: no PEM key")
	}
	var key *rsa.PrivateKey
	if k, err := x509.ParsePKCS8PrivateKey(blk.Bytes); err == nil {
		rk, ok := k.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("sso: the service provider key must be RSA")
		}
		key = rk
	} else if k, err := x509.ParsePKCS1PrivateKey(blk.Bytes); err == nil {
		key = k
	} else {
		return nil, errors.New("sso: unreadable key")
	}
	cert, err := parseCert(certB64)
	if err != nil {
		return nil, err
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok || pub.N.Cmp(key.N) != 0 {
		return nil, errors.New("sso: the certificate is not the key's")
	}
	return &SPKeyPair{Key: key, Cert: cert}, nil
}

// ParseSPCert reads a certificate alone (PEM or base64 DER), with no key: what an IdP admin is given.
func ParseSPCert(cert string) (*SPKeyPair, error) {
	c, err := parseCert(cert)
	if err != nil {
		return nil, err
	}
	return &SPKeyPair{Cert: c}, nil
}
