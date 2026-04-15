package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

type Manager struct {
	paths *system.Paths
}

func NewManager(paths *system.Paths) *Manager {
	return &Manager{paths: paths}
}

func (m *Manager) IsCAInstalled() bool {
	_, err1 := os.Stat(m.paths.CACert())
	_, err2 := os.Stat(m.paths.CAKey())
	return err1 == nil && err2 == nil
}

func (m *Manager) SetupCA() error {
	if m.IsCAInstalled() {
		return nil
	}

	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate template
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"LamboServer Local CA"},
			CommonName:   "LamboServer Local CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
	}

	// Self-sign the CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Write CA cert
	certFile, err := os.Create(m.paths.CACert())
	if err != nil {
		return err
	}
	defer certFile.Close()
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes})

	// Write CA key
	keyFile, err := os.OpenFile(m.paths.CAKey(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyFile.Close()
	pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)})

	return nil
}

func (m *Manager) TrustCA() error {
	if !m.IsCAInstalled() {
		return fmt.Errorf("CA not installed, run SetupCA first")
	}

	cmd := fmt.Sprintf("security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s", m.paths.CACert())
	return system.RunWithAdminPrivileges(cmd)
}

func (m *Manager) GenerateCert(domain string) (string, string, error) {
	certPath := m.paths.SiteCert(domain)
	keyPath := m.paths.SiteKey(domain)

	// Load CA
	caCertPEM, err := os.ReadFile(m.paths.CACert())
	if err != nil {
		return "", "", fmt.Errorf("failed to read CA cert: %w", err)
	}
	caKeyPEM, err := os.ReadFile(m.paths.CAKey())
	if err != nil {
		return "", "", fmt.Errorf("failed to read CA key: %w", err)
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA cert: %w", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA key: %w", err)
	}

	// Generate site key
	siteKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Create site certificate
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	site := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"LamboServer"},
			CommonName:   domain,
		},
		DNSNames:    []string{domain},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(2, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	siteBytes, err := x509.CreateCertificate(rand.Reader, site, caCert, &siteKey.PublicKey, caKey)
	if err != nil {
		return "", "", err
	}

	// Write site cert
	cf, err := os.Create(certPath)
	if err != nil {
		return "", "", err
	}
	defer cf.Close()
	pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: siteBytes})

	// Write site key
	kf, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", err
	}
	defer kf.Close()
	pem.Encode(kf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(siteKey)})

	return certPath, keyPath, nil
}

func (m *Manager) RemoveCert(domain string) error {
	os.Remove(m.paths.SiteCert(domain))
	os.Remove(m.paths.SiteKey(domain))
	return nil
}
