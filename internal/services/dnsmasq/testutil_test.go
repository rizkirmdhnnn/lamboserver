package dnsmasq_test

import (
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// --- MockFileSystem ---

// MockFileSystem implements dnsmasq.FileSystem for testing.
type MockFileSystem struct {
	StatFunc      func(name string) (fs.FileInfo, error)
	WriteFileFunc func(name string, data []byte, perm fs.FileMode) error
	ReadFileFunc  func(name string) ([]byte, error)
	MkdirAllFunc  func(name string, perm fs.FileMode) error

	// Call tracking
	StatCalls      []string
	WriteFileCalls []WriteFileCall
	ReadFileCalls  []string
	MkdirAllCalls  []string
}

type WriteFileCall struct {
	Name string
	Data []byte
	Perm fs.FileMode
}

func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	m.StatCalls = append(m.StatCalls, name)
	if m.StatFunc != nil {
		return m.StatFunc(name)
	}
	return &mockFileInfo{name: filepath.Base(name)}, nil
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	m.WriteFileCalls = append(m.WriteFileCalls, WriteFileCall{Name: name, Data: data, Perm: perm})
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(name, data, perm)
	}
	return nil
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	m.ReadFileCalls = append(m.ReadFileCalls, name)
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(name)
	}
	return nil, nil
}

func (m *MockFileSystem) MkdirAll(name string, perm fs.FileMode) error {
	m.MkdirAllCalls = append(m.MkdirAllCalls, name)
	if m.MkdirAllFunc != nil {
		return m.MkdirAllFunc(name, perm)
	}
	return nil
}

// mockFileInfo is a minimal fs.FileInfo implementation.
type mockFileInfo struct {
	name string
}

func (fi *mockFileInfo) Name() string       { return fi.name }
func (fi *mockFileInfo) Size() int64        { return 0 }
func (fi *mockFileInfo) Mode() fs.FileMode  { return 0644 }
func (fi *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *mockFileInfo) IsDir() bool        { return false }
func (fi *mockFileInfo) Sys() any           { return nil }

// --- MockLaunchdService ---

// MockLaunchdService implements dnsmasq.LaunchdService for testing.
type MockLaunchdService struct {
	InstallFunc   func(cfg system.ServiceConfig) error
	UninstallFunc func(cfg system.ServiceConfig) error
	IsRunningFunc func(label string) bool

	// Call tracking
	InstallCalls   []system.ServiceConfig
	UninstallCalls []system.ServiceConfig
	IsRunningCalls []string
}

func (m *MockLaunchdService) Install(cfg system.ServiceConfig) error {
	m.InstallCalls = append(m.InstallCalls, cfg)
	if m.InstallFunc != nil {
		return m.InstallFunc(cfg)
	}
	return nil
}

func (m *MockLaunchdService) Uninstall(cfg system.ServiceConfig) error {
	m.UninstallCalls = append(m.UninstallCalls, cfg)
	if m.UninstallFunc != nil {
		return m.UninstallFunc(cfg)
	}
	return nil
}

func (m *MockLaunchdService) IsRunning(label string) bool {
	m.IsRunningCalls = append(m.IsRunningCalls, label)
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc(label)
	}
	return false
}

// --- MockAdminRunner ---

// MockAdminRunner implements dnsmasq.AdminRunner for testing.
type MockAdminRunner struct {
	RunWithPrivilegesFunc func(command string) error

	// Call tracking
	RunWithPrivilegesCalls []string
}

func (m *MockAdminRunner) RunWithPrivileges(command string) error {
	m.RunWithPrivilegesCalls = append(m.RunWithPrivilegesCalls, command)
	if m.RunWithPrivilegesFunc != nil {
		return m.RunWithPrivilegesFunc(command)
	}
	return nil
}

// --- Test helpers ---

func newTestPaths(t *testing.T) *system.Paths {
	t.Helper()
	return &system.Paths{Home: t.TempDir()}
}
