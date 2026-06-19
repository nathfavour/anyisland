//go:build !linux

package cli

import (
	"fmt"
	"os"
)

func (b *UpdateBroker) getPeerPID(f *os.File) (int, error) {
	return 0, fmt.Errorf("peer PID resolution not supported on this platform")
}
