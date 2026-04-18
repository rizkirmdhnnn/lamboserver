package system_test

import (
	"path/filepath"
	"testing"

	"github.com/rizkirmdhnnn/lamboserver/internal/system"
)

// newTestPaths returns a Paths struct rooted at the given directory.
// Caller is responsible for providing a valid root (e.g. t.TempDir()).
func newTestPaths(root string) *system.Paths {
	return &system.Paths{Home: root}
}

// newTestIntegration creates an Integration wired to a temp home directory.
// Returns the integration and the temp home path.
// The temp dir is automatically cleaned up when the test ends.
func newTestIntegration(t *testing.T) (*system.Integration, string) {
	t.Helper()
	tmpHome := t.TempDir()
	paths := &system.Paths{Home: filepath.Join(tmpHome, ".lamboserver")}
	integration := system.NewIntegration(paths).WithHome(tmpHome)
	return integration, tmpHome
}
