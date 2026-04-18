package cert_test

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rizkirmdhnnn/lamboserver/internal/cert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestManager_IsCAInstalled_False(t *testing.T) {
	mgr, _, _ := newTestCertManager(t)
	assert.False(t, mgr.IsCAInstalled())
}

func TestManager_IsCAInstalled_OnlyCert(t *testing.T) {
	mgr, paths, _ := newTestCertManager(t)

	// Create only the CA cert file, not the key
	f, err := os.Create(paths.CACert())
	require.NoError(t, err)
	f.Close()

	assert.False(t, mgr.IsCAInstalled(), "should return false when only cert exists, key is missing")
}

func TestManager_IsCAInstalled_True(t *testing.T) {
	mgr, _, _ := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())
	assert.True(t, mgr.IsCAInstalled())
}

func TestManager_SetupCA_GeneratesFiles(t *testing.T) {
	mgr, paths, _ := newTestCertManager(t)

	err := mgr.SetupCA()
	require.NoError(t, err)

	// Check cert file exists and starts with correct PEM header
	certData, err := os.ReadFile(paths.CACert())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(certData), "-----BEGIN CERTIFICATE-----"), "CA cert should start with BEGIN CERTIFICATE")

	// Check key file exists and starts with correct PEM header
	keyData, err := os.ReadFile(paths.CAKey())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(keyData), "-----BEGIN RSA PRIVATE KEY-----"), "CA key should start with BEGIN RSA PRIVATE KEY")
}

func TestManager_SetupCA_Idempotent(t *testing.T) {
	mgr, paths, _ := newTestCertManager(t)

	require.NoError(t, mgr.SetupCA())

	// Capture file contents after first call
	certData1, err := os.ReadFile(paths.CACert())
	require.NoError(t, err)
	keyData1, err := os.ReadFile(paths.CAKey())
	require.NoError(t, err)

	// Second call should be a no-op
	require.NoError(t, mgr.SetupCA())

	certData2, err := os.ReadFile(paths.CACert())
	require.NoError(t, err)
	keyData2, err := os.ReadFile(paths.CAKey())
	require.NoError(t, err)

	assert.Equal(t, certData1, certData2, "CA cert should be unchanged on second SetupCA call")
	assert.Equal(t, keyData1, keyData2, "CA key should be unchanged on second SetupCA call")
}

func TestManager_SetupCA_CertIsValidCA(t *testing.T) {
	mgr, paths, _ := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())

	certData, err := os.ReadFile(paths.CACert())
	require.NoError(t, err)

	block, _ := pem.Decode(certData)
	require.NotNil(t, block, "PEM block should not be nil")

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	assert.True(t, cert.IsCA, "certificate should be a CA")
	assert.Equal(t, "LamboServer Local CA", cert.Subject.CommonName)

	expectedExpiry := time.Now().AddDate(10, 0, 0)
	// Allow a few seconds tolerance
	assert.WithinDuration(t, expectedExpiry, cert.NotAfter, 10*time.Second)
}

func TestManager_TrustCA_Success(t *testing.T) {
	mgr, paths, adminMock := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())

	adminMock.On("RunWithPrivileges", mock.MatchedBy(func(cmd string) bool {
		return strings.Contains(cmd, "security add-trusted-cert") && strings.Contains(cmd, paths.CACert())
	})).Return(nil)

	err := mgr.TrustCA()
	require.NoError(t, err)

	adminMock.AssertExpectations(t)
}

func TestManager_TrustCA_NoCA(t *testing.T) {
	mgr, _, _ := newTestCertManager(t)

	err := mgr.TrustCA()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CA not installed")
}

func TestManager_TrustCA_AdminFails(t *testing.T) {
	mgr, _, adminMock := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())

	expectedErr := errors.New("admin privileges denied")
	adminMock.On("RunWithPrivileges", mock.Anything).Return(expectedErr)

	err := mgr.TrustCA()
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestManager_GenerateCert_Success(t *testing.T) {
	mgr, paths, adminMock := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())

	adminMock.On("RunWithPrivileges", mock.Anything).Return(nil)

	certPath, keyPath, err := mgr.GenerateCert("myapp.test")
	require.NoError(t, err)
	assert.NotEmpty(t, certPath)
	assert.NotEmpty(t, keyPath)

	// Verify files exist
	_, err = os.Stat(certPath)
	assert.NoError(t, err, "cert file should exist")
	_, err = os.Stat(keyPath)
	assert.NoError(t, err, "key file should exist")

	// Verify cert PEM content
	certData, err := os.ReadFile(certPath)
	require.NoError(t, err)

	block, _ := pem.Decode(certData)
	require.NotNil(t, block)

	siteCert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	assert.Equal(t, "myapp.test", siteCert.Subject.CommonName)
	assert.Contains(t, siteCert.DNSNames, "myapp.test")

	// Verify issuer matches CA
	caCertData, err := os.ReadFile(paths.CACert())
	require.NoError(t, err)
	caBlock, _ := pem.Decode(caCertData)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	require.NoError(t, err)

	assert.Equal(t, caCert.Subject.CommonName, siteCert.Issuer.CommonName)
}

func TestManager_GenerateCert_NoCA(t *testing.T) {
	mgr, _, _ := newTestCertManager(t)

	_, _, err := mgr.GenerateCert("myapp.test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read CA cert")
}

func TestManager_RemoveCert(t *testing.T) {
	mgr, paths, adminMock := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())

	adminMock.On("RunWithPrivileges", mock.Anything).Return(nil)

	_, _, err := mgr.GenerateCert("myapp.test")
	require.NoError(t, err)

	// Verify cert files exist
	_, err = os.Stat(paths.SiteCert("myapp.test"))
	require.NoError(t, err)
	_, err = os.Stat(paths.SiteKey("myapp.test"))
	require.NoError(t, err)

	// Remove cert
	err = mgr.RemoveCert("myapp.test")
	require.NoError(t, err)

	// Verify cert files no longer exist
	_, err = os.Stat(paths.SiteCert("myapp.test"))
	assert.True(t, os.IsNotExist(err), "cert file should no longer exist")
	_, err = os.Stat(paths.SiteKey("myapp.test"))
	assert.True(t, os.IsNotExist(err), "key file should no longer exist")
}

func TestManager_RemoveCert_NonexistentDomain(t *testing.T) {
	mgr, _, _ := newTestCertManager(t)

	// Should not error even when files don't exist
	err := mgr.RemoveCert("nope.test")
	assert.NoError(t, err)
}

func TestManager_TrustCA_ShellQuotesCACertPath(t *testing.T) {
	mgr, paths, adminMock := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())

	adminMock.On("RunWithPrivileges", mock.MatchedBy(func(cmd string) bool {
		return strings.Contains(cmd, "security add-trusted-cert") &&
			strings.Contains(cmd, "'"+paths.CACert()+"'")
	})).Return(nil)

	err := mgr.TrustCA()
	require.NoError(t, err)
	adminMock.AssertExpectations(t)
}

func TestManager_GenerateCert_CorruptedCACert(t *testing.T) {
	mgr, paths, _ := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())

	// Overwrite the CA cert file with invalid PEM data
	require.NoError(t, os.WriteFile(paths.CACert(), []byte("not valid pem data"), 0644))

	_, _, err := mgr.GenerateCert("test.test")
	require.Error(t, err)
	assert.True(t, errors.Is(err, cert.ErrCorruptedCA), "expected ErrCorruptedCA, got: %v", err)
}

func TestManager_GenerateCert_CorruptedCAKey(t *testing.T) {
	mgr, paths, _ := newTestCertManager(t)
	require.NoError(t, mgr.SetupCA())

	// Overwrite the CA key file with invalid PEM data
	require.NoError(t, os.WriteFile(paths.CAKey(), []byte("not valid pem data"), 0600))

	_, _, err := mgr.GenerateCert("test.test")
	require.Error(t, err)
	assert.True(t, errors.Is(err, cert.ErrCorruptedCA), "expected ErrCorruptedCA, got: %v", err)
}
