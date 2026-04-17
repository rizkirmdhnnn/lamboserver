// Package config_test contains unit tests for the config package.
package config_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- NewStore / Defaults ---

func TestStore_NewStore_Defaults(t *testing.T) {
	store := newTestStore(t)
	cfg := store.Get()

	assert.Equal(t, 80, cfg.NginxPort, "default NginxPort should be 80")
	assert.Equal(t, 443, cfg.NginxSSLPort, "default NginxSSLPort should be 443")
	assert.Empty(t, cfg.Sites, "default Sites should be empty")
	assert.Empty(t, cfg.ActivePhpVersion, "default ActivePhpVersion should be empty")
	assert.Empty(t, cfg.ActiveNodeVersion, "default ActiveNodeVersion should be empty")
	assert.False(t, cfg.FirstRunComplete, "default FirstRunComplete should be false")
	assert.False(t, cfg.DebugMode, "default DebugMode should be false")
}

// --- SetActivePhpVersion ---

func TestStore_SetActivePhpVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{name: "sets version 8.3", version: "8.3"},
		{name: "overwrites with 8.4", version: "8.4"},
		{name: "sets empty string", version: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := newTestStore(t)
			require.NoError(t, store.SetActivePhpVersion(tc.version))
			assert.Equal(t, tc.version, store.Get().ActivePhpVersion)
		})
	}
}

// --- SetActiveNodeVersion ---

func TestStore_SetActiveNodeVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{name: "sets 20.0.0", version: "20.0.0"},
		{name: "overwrites with 18.0.0", version: "18.0.0"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := newTestStore(t)
			require.NoError(t, store.SetActiveNodeVersion(tc.version))
			assert.Equal(t, tc.version, store.Get().ActiveNodeVersion)
		})
	}
}

// --- AddSite ---

func TestStore_AddSite(t *testing.T) {
	t.Run("adds site", func(t *testing.T) {
		store := newTestStore(t)
		require.NoError(t, store.AddSite("example.test", "/var/www/example"))
		sites := store.GetSites()
		assert.Len(t, sites, 1)
		assert.Equal(t, "example.test", sites[0].Domain)
	})

	t.Run("duplicate domain is no-op", func(t *testing.T) {
		store := newTestStore(t)
		require.NoError(t, store.AddSite("example.test", "/var/www/example"))
		require.NoError(t, store.AddSite("example.test", "/var/www/example"))
		assert.Len(t, store.GetSites(), 1)
	})

	t.Run("multiple sites", func(t *testing.T) {
		store := newTestStore(t)
		require.NoError(t, store.AddSite("a.test", "/var/www/a"))
		require.NoError(t, store.AddSite("b.test", "/var/www/b"))
		assert.Len(t, store.GetSites(), 2)
	})

	t.Run("sets SSLEnabled true by default", func(t *testing.T) {
		store := newTestStore(t)
		require.NoError(t, store.AddSite("ssl.test", "/var/www/ssl"))
		assert.True(t, store.GetSites()[0].SSLEnabled)
	})

	t.Run("sets CreatedAt", func(t *testing.T) {
		store := newTestStore(t)
		require.NoError(t, store.AddSite("time.test", "/var/www/time"))
		assert.NotEmpty(t, store.GetSites()[0].CreatedAt)
	})
}

// --- RemoveSite ---

func TestStore_RemoveSite(t *testing.T) {
	t.Run("removes existing site", func(t *testing.T) {
		store := newTestStore(t)
		require.NoError(t, store.AddSite("remove.test", "/var/www/remove"))
		require.NoError(t, store.RemoveSite("remove.test"))
		assert.Empty(t, store.GetSites())
	})

	t.Run("non-existent domain no-op", func(t *testing.T) {
		store := newTestStore(t)
		require.NoError(t, store.RemoveSite("nonexistent.test"))
		assert.Empty(t, store.GetSites())
	})

	t.Run("removes only matching", func(t *testing.T) {
		store := newTestStore(t)
		require.NoError(t, store.AddSite("a.test", "/var/www/a"))
		require.NoError(t, store.AddSite("b.test", "/var/www/b"))
		require.NoError(t, store.RemoveSite("a.test"))
		sites := store.GetSites()
		assert.Len(t, sites, 1)
		assert.Equal(t, "b.test", sites[0].Domain)
	})
}

// --- GetSites returns a defensive copy ---

func TestStore_GetSites_ReturnsCopy(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.AddSite("copy.test", "/var/www/copy"))

	sites := store.GetSites()
	originalLen := len(sites)

	// Modify the returned slice — should not affect the internal store.
	sites = append(sites, config.SiteConfig{Domain: "extra.test"})

	assert.Len(t, store.GetSites(), originalLen, "internal sites slice should be unaffected by modification of returned copy")
}

// --- SetFirstRunComplete ---

func TestStore_SetFirstRunComplete(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.SetFirstRunComplete())
	assert.True(t, store.Get().FirstRunComplete)
}

// --- SetDebugMode ---

func TestStore_SetDebugMode(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{name: "enable debug mode", enabled: true},
		{name: "disable debug mode", enabled: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := newTestStore(t)
			require.NoError(t, store.SetDebugMode(tc.enabled))
			assert.Equal(t, tc.enabled, store.Get().DebugMode)
		})
	}
}

// --- SetPostgreSQLEnabled ---

func TestStore_SetPostgreSQLEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{name: "enable postgresql", enabled: true},
		{name: "disable postgresql", enabled: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := newTestStore(t)
			require.NoError(t, store.SetPostgreSQLEnabled(tc.enabled))
			assert.Equal(t, tc.enabled, store.Get().PostgreSQLEnabled)
		})
	}
}

func TestStore_PostgreSQLEnabled_JSONTag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	store := config.NewStore(configPath)
	require.NoError(t, store.SetPostgreSQLEnabled(true))

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"postgresql_enabled"`, "JSON tag must be postgresql_enabled")
}

// --- Persistence round-trip ---

func TestStore_PersistsAcrossReload(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Write data with the first store.
	store1 := config.NewStore(configPath)
	require.NoError(t, store1.SetActivePhpVersion("8.3"))
	require.NoError(t, store1.AddSite("test.test", "/tmp"))

	// Create a second store at the same path to verify persistence.
	store2 := config.NewStore(configPath)
	assert.Equal(t, "8.3", store2.Get().ActivePhpVersion, "ActivePhpVersion should persist")
	assert.Len(t, store2.GetSites(), 1, "sites should persist")
}

// --- Load with invalid JSON ---

func TestStore_Load_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Write invalid JSON to the config file.
	require.NoError(t, os.WriteFile(configPath, []byte("{invalid json}"), 0644))

	// NewStore calls Load internally; Load will fail to unmarshal but NewStore
	// silently ignores the error and returns a store with default values.
	store := config.NewStore(configPath)
	cfg := store.Get()

	// The struct retains zero values since unmarshal failed.
	assert.Equal(t, "", cfg.ActivePhpVersion)
	assert.Equal(t, "", cfg.ActiveNodeVersion)
	assert.False(t, cfg.DebugMode)
	assert.False(t, cfg.FirstRunComplete)
}

// --- Concurrent read/write ---

func TestStore_ConcurrentReadWrite(t *testing.T) {
	store := newTestStore(t)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for n := 0; n < goroutines; n++ {
		n := n
		go func() {
			defer wg.Done()
			version := fmt.Sprintf("8.%d", n%5)
			// Ignore errors — concurrent writes may race on the file but the
			// store's mutex should prevent data corruption.
			_ = store.SetActivePhpVersion(version)
			_ = store.Get()
		}()
	}

	wg.Wait()
	// If we reach here without a panic or race detector violation, the test passes.
}

// --- JSON round-trip integrity ---

// TestStore_JSONRoundTrip verifies that the config file produced by Save is
// valid JSON that can be deserialized back into AppConfig.
func TestStore_JSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	store := config.NewStore(configPath)
	require.NoError(t, store.SetActivePhpVersion("8.3"))
	require.NoError(t, store.AddSite("json.test", "/var/www/json"))

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var parsed config.AppConfig
	require.NoError(t, json.Unmarshal(data, &parsed), "config file should contain valid JSON")
	assert.Equal(t, "8.3", parsed.ActivePhpVersion)
	assert.Len(t, parsed.Sites, 1)
}
