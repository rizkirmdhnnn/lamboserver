package process

import (
	"fmt"
	"net"
)

// IsPortAvailable checks whether the given TCP port is free to bind on localhost.
func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
