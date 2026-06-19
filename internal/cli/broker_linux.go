//go:build linux

package cli

import (
	"os"
	"syscall"
)

func (b *UpdateBroker) getPeerPID(f *os.File) (int, error) {
	ucred, err := syscall.GetsockoptUcred(int(f.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return 0, err
	}
	return int(ucred.Pid), nil
}
