// Package notify provides desktop notification support for LamboServer.
// It sends native OS notifications for service state changes, installation
// completions, and error alerts.
package notify

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Send displays a native desktop notification with the given title and message.
func Send(title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, message, title)
		return exec.Command("osascript", "-e", script).Run()
	default:
		// Placeholder for Linux (notify-send) and Windows (toast) support.
		return nil
	}
}
