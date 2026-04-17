// Package config_test provides test helpers for the config package.
package config_test

import (
	"path/filepath"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/config"
	"github.com/stretchr/testify/require"
)

// newTestStore creates a Store backed by a temporary directory.
// The config file is placed at <t.TempDir()>/config.json.
// The temp dir is automatically cleaned up when the test completes.
func newTestStore(t *testing.T) *config.Store {
	t.Helper()
	dir := t.TempDir()
	return config.NewStore(filepath.Join(dir, "config.json"))
}

// newTestStoreWithData creates a Store pre-populated with the given sites.
// Each site is added via AddSite using the site's Domain and Path fields.
func newTestStoreWithData(t *testing.T, sites []config.SiteConfig) *config.Store {
	t.Helper()
	store := newTestStore(t)
	for _, s := range sites {
		require.NoError(t, store.AddSite(s.Domain, s.Path))
	}
	return store
}
