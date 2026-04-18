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
	"path/filepath"
	"strings"
	"time"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// Manager handles local CA creation, keychain trust, and per-site TLS certificate
// generation. All cryptographic operations use the standard library (crypto/rsa,
// crypto/x509); no external CA tooling is required.
type Manager struct {
	paths *system.Paths
	fs    FileSystem
	admin AdminRunner
}

// NewManager creates a Manager with injected filesystem and admin-runner dependencies.
// paths provides the CA and certificate file locations under ~/.lamboserver/certs/.
func NewManager(paths *system.Paths, fs FileSystem, admin AdminRunner) *Manager {
	return &Manager{paths: paths, fs: fs, admin: admin}
}

// IsCAInstalled reports whether both the CA certificate and CA private key files exist
// at the paths returned by paths.CACert() and paths.CAKey().
func (m *Manager) IsCAInstalled() bool {
	_, err1 := m.fs.Stat(m.paths.CACert())
	_, err2 := m.fs.Stat(m.paths.CAKey())
	return err1 == nil && err2 == nil
}

// SetupCA generates a new local Certificate Authority: a 4096-bit RSA private key and
// a self-signed CA certificate valid for 10 years. Both files are written to
// ~/.lamboserver/certs/. If the CA is already installed, SetupCA is a no-op.
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
	certFile, err := m.fs.Create(m.paths.CACert())
	if err != nil {
		return err
	}
	defer certFile.Close()
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}) // error intentionally ignored: pem.Encode to file rarely fails

	// Write CA key
	keyFile, err := m.fs.OpenFile(m.paths.CAKey(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyFile.Close()
	pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)}) // error intentionally ignored: pem.Encode to file rarely fails

	return nil
}

// TrustCA adds the local CA certificate to the macOS System keychain as a trusted root
// using the `security` CLI via AdminRunner (requires admin privileges). Returns an error
// if the CA is not yet installed.
func (m *Manager) TrustCA() error {
	if !m.IsCAInstalled() {
		return fmt.Errorf("CA not installed, run SetupCA first")
	}

	caPath := filepath.Clean(m.paths.CACert())
	if strings.ContainsAny(caPath, ";&|`$><\\\"'\n\r\t") {
		return fmt.Errorf("invalid CA certificate path: %q", caPath)
	}
	cmd := fmt.Sprintf("security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '%s'", caPath)
	return m.admin.RunWithPrivileges(cmd)
}

// GenerateCert generates a 2048-bit RSA TLS certificate for the given domain, signed
// by the local CA. The certificate is valid for 2 years with SAN for the exact domain.
// Returns the paths to the certificate and private key files, or an error if the CA
// is not installed or cryptographic operations fail.
func (m *Manager) GenerateCert(domain string) (string, string, error) {
	certPath := m.paths.SiteCert(domain)
	keyPath := m.paths.SiteKey(domain)

	// Load CA
	caCertPEM, err := m.fs.ReadFile(m.paths.CACert())
	if err != nil {
		return "", "", fmt.Errorf("failed to read CA cert: %w", err)
	}
	caKeyPEM, err := m.fs.ReadFile(m.paths.CAKey())
	if err != nil {
		return "", "", fmt.Errorf("failed to read CA key: %w", err)
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return "", "", ErrCorruptedCA
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA cert: %w", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return "", "", ErrCorruptedCA
	}
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
	cf, err := m.fs.Create(certPath)
	if err != nil {
		return "", "", err
	}
	defer cf.Close()
	pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: siteBytes}) // error intentionally ignored: pem.Encode to file rarely fails

	// Write site key
	kf, err := m.fs.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", err
	}
	defer kf.Close()
	pem.Encode(kf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(siteKey)}) // error intentionally ignored: pem.Encode to file rarely fails

	return certPath, keyPath, nil
}

// RemoveCert deletes the certificate and key files for the given domain from
// ~/.lamboserver/certs/. Filesystem errors are intentionally ignored during cleanup.
func (m *Manager) RemoveCert(domain string) error {
	m.fs.Remove(m.paths.SiteCert(domain)) // error intentionally ignored on cleanup
	m.fs.Remove(m.paths.SiteKey(domain))  // error intentionally ignored on cleanup
	return nil
}
