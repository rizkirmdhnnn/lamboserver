// Package integration_test contains integration tests that wire together real manager
// implementations with lightweight mock adapters (no real launchd, no real filesystem
// mutations outside t.TempDir).
package integration_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/dnsmasq"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/nginx"
	"github.com/rizkirmdhnnn/lamboserver/internal/services/php"
	"github.com/rizkirmdhnnn/lamboserver/internal/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// RecordingLaunchdService — call-capture adapter (D-03)
// ---------------------------------------------------------------------------

// CallRecord captures one Install/Uninstall/IsRunning invocation.
type CallRecord struct {
	Method    string    // "Install", "Uninstall", "IsRunning"
	Label     string    // service label
	Timestamp time.Time // wall-clock time of the call
	Seq       int       // monotonic sequence number (1-based)
}

// RecordingLaunchdService records every launchd call without touching the OS.
// It satisfies nginx.LaunchdService, dnsmasq.LaunchdService, and php.LaunchdService
// (all three interfaces share the same method set).
type RecordingLaunchdService struct {
	mu           sync.Mutex
	calls        []CallRecord
	seq          int
	installErr   error
	uninstallErr error
	running      map[string]bool // label -> simulated running state
}

func newRecordingLaunchd() *RecordingLaunchdService {
	return &RecordingLaunchdService{running: make(map[string]bool)}
}

func (r *RecordingLaunchdService) record(method, label string) {
	r.seq++
	r.calls = append(r.calls, CallRecord{
		Method:    method,
		Label:     label,
		Timestamp: time.Now(),
		Seq:       r.seq,
	})
}

// Install records the call, auto-marks the service as running, returns installErr.
func (r *RecordingLaunchdService) Install(cfg system.ServiceConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.record("Install", cfg.Label)
	if r.installErr == nil {
		r.running[cfg.Label] = true
	}
	return r.installErr
}

// Uninstall records the call, marks the service as not running, returns uninstallErr.
func (r *RecordingLaunchdService) Uninstall(cfg system.ServiceConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.record("Uninstall", cfg.Label)
	if r.uninstallErr == nil {
		r.running[cfg.Label] = false
	}
	return r.uninstallErr
}

// IsRunning records the call and returns the simulated running state.
func (r *RecordingLaunchdService) IsRunning(label string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.record("IsRunning", label)
	return r.running[label]
}

// Calls returns a copy of all recorded calls.
func (r *RecordingLaunchdService) Calls() []CallRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]CallRecord, len(r.calls))
	copy(out, r.calls)
	return out
}

// InstallCalls returns only Install calls in recorded order.
func (r *RecordingLaunchdService) InstallCalls() []CallRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []CallRecord
	for _, c := range r.calls {
		if c.Method == "Install" {
			out = append(out, c)
		}
	}
	return out
}

// UninstallCalls returns only Uninstall calls in recorded order.
func (r *RecordingLaunchdService) UninstallCalls() []CallRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []CallRecord
	for _, c := range r.calls {
		if c.Method == "Uninstall" {
			out = append(out, c)
		}
	}
	return out
}

// CallsForLabel returns all calls for a specific service label.
func (r *RecordingLaunchdService) CallsForLabel(label string) []CallRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []CallRecord
	for _, c := range r.calls {
		if c.Label == label {
			out = append(out, c)
		}
	}
	return out
}

// Compile-time interface checks — RecordingLaunchdService satisfies all three interfaces.
var _ nginx.LaunchdService = (*RecordingLaunchdService)(nil)
var _ dnsmasq.LaunchdService = (*RecordingLaunchdService)(nil)
var _ php.LaunchdService = (*RecordingLaunchdService)(nil)

// ---------------------------------------------------------------------------
// MockNginxFS — real filesystem ops rooted in t.TempDir()
// ---------------------------------------------------------------------------

// MockNginxFS implements nginx.FileSystem using the real os package.
// All operations work against the real filesystem; test directories are
// created under t.TempDir() so they are cleaned up automatically.
type MockNginxFS struct{}

func (MockNginxFS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (MockNginxFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}
func (MockNginxFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

var _ nginx.FileSystem = MockNginxFS{}

// ---------------------------------------------------------------------------
// MockNginxHelperRunner — no-op helper (nginx.HelperRunner)
// ---------------------------------------------------------------------------

// MockNginxHelperRunner is a no-op implementation of nginx.HelperRunner.
// Nginx.Manager uses it for reload operations; for lifecycle tests we just
// need it to return nil.
type MockNginxHelperRunner struct{}

func (MockNginxHelperRunner) Run(args ...string) (string, error) { return "", nil }

var _ nginx.HelperRunner = MockNginxHelperRunner{}

// ---------------------------------------------------------------------------
// MockDnsFS — real filesystem ops for dnsmasq.Manager
// ---------------------------------------------------------------------------

// MockDnsFS implements dnsmasq.FileSystem using the real os package.
type MockDnsFS struct{}

func (MockDnsFS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (MockDnsFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}
func (MockDnsFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
func (MockDnsFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

var _ dnsmasq.FileSystem = MockDnsFS{}

// ---------------------------------------------------------------------------
// MockDnsAdminRunner — no-op privilege runner (dnsmasq.AdminRunner)
// ---------------------------------------------------------------------------

// MockDnsAdminRunner implements dnsmasq.AdminRunner; RunWithPrivileges always returns nil.
type MockDnsAdminRunner struct{}

func (MockDnsAdminRunner) RunWithPrivileges(command string) error { return nil }

var _ dnsmasq.AdminRunner = MockDnsAdminRunner{}

// ---------------------------------------------------------------------------
// MockPhpFS — real filesystem ops for php.FpmManager
// ---------------------------------------------------------------------------

// MockPhpFS implements php.FileSystem using the real os package.
type MockPhpFS struct{}

func (MockPhpFS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (MockPhpFS) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }
func (MockPhpFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
func (MockPhpFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}
func (MockPhpFS) Remove(name string) error           { return os.Remove(name) }
func (MockPhpFS) RemoveAll(path string) error        { return os.RemoveAll(path) }
func (MockPhpFS) Chmod(name string, mode fs.FileMode) error { return os.Chmod(name, mode) }

var _ php.FileSystem = MockPhpFS{}

// ---------------------------------------------------------------------------
// Helpers (uses newTestPaths and newTestStore from helpers_test.go)
// ---------------------------------------------------------------------------

// createFakeExecutable writes an empty file at path and makes it executable.
func createFakeExecutable(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"), 0755))
}

// newTestStoreWithConfig creates a config.Store and sets the active PHP version if specified.
func newTestStoreWithConfig(t *testing.T, paths *system.Paths, cfg config.AppConfig) *config.Store {
	t.Helper()
	store := newTestStore(t, paths)
	if cfg.ActivePhpVersion != "" {
		require.NoError(t, store.SetActivePhpVersion(cfg.ActivePhpVersion))
	}
	return store
}

// ---------------------------------------------------------------------------
// TestServiceLifecycle — INTG-04
// ---------------------------------------------------------------------------

func TestServiceLifecycle(t *testing.T) {
	t.Run("NginxStartStop", func(t *testing.T) {
		paths := newTestPaths(t)
		recorder := newRecordingLaunchd()
		// Create a fake nginx binary so IsInstalled() returns true.
		createFakeExecutable(t, paths.NginxBin())

		mgr := nginx.NewManager(paths, recorder, MockNginxFS{}, MockNginxHelperRunner{})

		require.NoError(t, mgr.Start(), "Start() should succeed")

		installs := recorder.InstallCalls()
		require.Len(t, installs, 1, "expected exactly 1 Install call after Start")
		assert.Equal(t, nginx.ServiceLabel, installs[0].Label)

		require.NoError(t, mgr.Stop(), "Stop() should succeed")

		uninstalls := recorder.UninstallCalls()
		require.Len(t, uninstalls, 1, "expected exactly 1 Uninstall call after Stop")
		assert.Equal(t, nginx.ServiceLabel, uninstalls[0].Label)
	})

	t.Run("PhpFpmStartStop", func(t *testing.T) {
		paths := newTestPaths(t)
		recorder := newRecordingLaunchd()

		// Create a fake php-fpm binary at the expected location.
		fpmBin := filepath.Join(paths.PhpVersionDir("8.3"), "sbin", "php-fpm")
		createFakeExecutable(t, fpmBin)

		store := newTestStoreWithConfig(t, paths, config.AppConfig{ActivePhpVersion: "8.3"})
		mgr := php.NewFpmManager(paths, store, recorder, MockPhpFS{})

		require.NoError(t, mgr.Start(), "FpmManager.Start() should succeed")

		installs := recorder.InstallCalls()
		require.Len(t, installs, 1, "expected exactly 1 Install call after Start")
		assert.Equal(t, php.FpmServiceLabel, installs[0].Label)

		require.NoError(t, mgr.Stop(), "FpmManager.Stop() should succeed")

		uninstalls := recorder.UninstallCalls()
		require.Len(t, uninstalls, 1, "expected exactly 1 Uninstall call after Stop")
		assert.Equal(t, php.FpmServiceLabel, uninstalls[0].Label)
	})

	t.Run("DnsStartStop", func(t *testing.T) {
		paths := newTestPaths(t)
		recorder := newRecordingLaunchd()
		// Create a fake dnsmasq binary.
		createFakeExecutable(t, paths.DnsmasqBin())

		mgr := dnsmasq.NewManager(paths, recorder, MockDnsFS{}, MockDnsAdminRunner{})

		require.NoError(t, mgr.Start(), "dnsmasq.Manager.Start() should succeed")

		installs := recorder.InstallCalls()
		require.Len(t, installs, 1, "expected exactly 1 Install call after Start")
		assert.Equal(t, dnsmasq.ServiceLabel, installs[0].Label)

		require.NoError(t, mgr.Stop(), "dnsmasq.Manager.Stop() should succeed")

		uninstalls := recorder.UninstallCalls()
		require.Len(t, uninstalls, 1, "expected exactly 1 Uninstall call after Stop")
		assert.Equal(t, dnsmasq.ServiceLabel, uninstalls[0].Label)
	})

	t.Run("StartOrder", func(t *testing.T) {
		// Uses a single shared recorder to verify cross-service start ordering,
		// mirroring the ensureServicesRunning pattern in app.go.
		paths := newTestPaths(t)
		recorder := newRecordingLaunchd()

		// Create all required fake binaries.
		createFakeExecutable(t, paths.NginxBin())
		createFakeExecutable(t, paths.DnsmasqBin())
		fpmBin := filepath.Join(paths.PhpVersionDir("8.3"), "sbin", "php-fpm")
		createFakeExecutable(t, fpmBin)

		store := newTestStoreWithConfig(t, paths, config.AppConfig{ActivePhpVersion: "8.3"})

		nginxMgr := nginx.NewManager(paths, recorder, MockNginxFS{}, MockNginxHelperRunner{})
		dnsMgr := dnsmasq.NewManager(paths, recorder, MockDnsFS{}, MockDnsAdminRunner{})
		fpmMgr := php.NewFpmManager(paths, store, recorder, MockPhpFS{})

		// Start in app.go order: nginx -> dns -> php-fpm.
		require.NoError(t, nginxMgr.Start())
		require.NoError(t, dnsMgr.Start())
		require.NoError(t, fpmMgr.Start())

		installs := recorder.InstallCalls()
		require.Len(t, installs, 3, "expected 3 Install calls (one per service)")

		// Verify monotonically increasing sequence numbers.
		assert.Less(t, installs[0].Seq, installs[1].Seq, "nginx Install seq should precede dns Install seq")
		assert.Less(t, installs[1].Seq, installs[2].Seq, "dns Install seq should precede fpm Install seq")

		// Verify label order matches start order.
		assert.Equal(t, nginx.ServiceLabel, installs[0].Label, "first Install should be nginx")
		assert.Equal(t, dnsmasq.ServiceLabel, installs[1].Label, "second Install should be dnsmasq")
		assert.Equal(t, php.FpmServiceLabel, installs[2].Label, "third Install should be php-fpm")
	})

	t.Run("StopOrder", func(t *testing.T) {
		// Verifies that app.go shutdown order (fpm -> nginx -> dns) is the reverse of start.
		paths := newTestPaths(t)
		recorder := newRecordingLaunchd()

		createFakeExecutable(t, paths.NginxBin())
		createFakeExecutable(t, paths.DnsmasqBin())
		fpmBin := filepath.Join(paths.PhpVersionDir("8.3"), "sbin", "php-fpm")
		createFakeExecutable(t, fpmBin)

		store := newTestStoreWithConfig(t, paths, config.AppConfig{ActivePhpVersion: "8.3"})

		nginxMgr := nginx.NewManager(paths, recorder, MockNginxFS{}, MockNginxHelperRunner{})
		dnsMgr := dnsmasq.NewManager(paths, recorder, MockDnsFS{}, MockDnsAdminRunner{})
		fpmMgr := php.NewFpmManager(paths, store, recorder, MockPhpFS{})

		// Start all services first.
		require.NoError(t, nginxMgr.Start())
		require.NoError(t, dnsMgr.Start())
		require.NoError(t, fpmMgr.Start())

		// Stop in reverse order: php-fpm -> nginx -> dns (matching app.go shutdown).
		require.NoError(t, fpmMgr.Stop())
		require.NoError(t, nginxMgr.Stop())
		require.NoError(t, dnsMgr.Stop())

		uninstalls := recorder.UninstallCalls()
		require.Len(t, uninstalls, 3, "expected 3 Uninstall calls (one per service)")

		// Verify label order matches shutdown order.
		assert.Equal(t, php.FpmServiceLabel, uninstalls[0].Label, "first Uninstall should be php-fpm")
		assert.Equal(t, nginx.ServiceLabel, uninstalls[1].Label, "second Uninstall should be nginx")
		assert.Equal(t, dnsmasq.ServiceLabel, uninstalls[2].Label, "third Uninstall should be dnsmasq")

		// Verify monotonically increasing sequence numbers for Uninstall calls.
		assert.Less(t, uninstalls[0].Seq, uninstalls[1].Seq, "fpm Uninstall seq should precede nginx Uninstall seq")
		assert.Less(t, uninstalls[1].Seq, uninstalls[2].Seq, "nginx Uninstall seq should precede dns Uninstall seq")
	})

	t.Run("FpmRestartRecording", func(t *testing.T) {
		paths := newTestPaths(t)
		recorder := newRecordingLaunchd()

		fpmBin := filepath.Join(paths.PhpVersionDir("8.3"), "sbin", "php-fpm")
		createFakeExecutable(t, fpmBin)

		store := newTestStoreWithConfig(t, paths, config.AppConfig{ActivePhpVersion: "8.3"})
		mgr := php.NewFpmManager(paths, store, recorder, MockPhpFS{})

		// Start FPM first so it is in a "running" state.
		require.NoError(t, mgr.Start(), "initial Start() should succeed")

		// Restart = Stop then Start.
		require.NoError(t, mgr.Restart(), "Restart() should succeed")

		fpmCalls := recorder.CallsForLabel(php.FpmServiceLabel)

		// Collect Install and Uninstall calls in order for fpm label.
		var installs, uninstalls []CallRecord
		for _, c := range fpmCalls {
			switch c.Method {
			case "Install":
				installs = append(installs, c)
			case "Uninstall":
				uninstalls = append(uninstalls, c)
			}
		}

		// After Start + Restart: 2 Installs, 1 Uninstall from Restart.Stop().
		require.GreaterOrEqual(t, len(installs), 2, "expected at least 2 Install calls (Start + Restart.Start)")
		require.GreaterOrEqual(t, len(uninstalls), 1, "expected at least 1 Uninstall call (Restart.Stop)")

		// The Uninstall from Restart.Stop() must occur before the last Install of Restart.Start().
		lastUninstall := uninstalls[len(uninstalls)-1]
		lastInstall := installs[len(installs)-1]
		assert.Less(t, lastUninstall.Seq, lastInstall.Seq,
			"Uninstall seq (%d) should precede Install seq (%d) within Restart",
			lastUninstall.Seq, lastInstall.Seq)
	})
}
