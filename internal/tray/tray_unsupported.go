//go:build !darwin

package tray

// Start is a no-op on non-darwin platforms.
func (c *Controller) Start() {}

// Destroy is a no-op on non-darwin platforms.
func (c *Controller) Destroy() {}
