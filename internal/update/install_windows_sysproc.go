//go:build windows

package update

import "syscall"

func windowsDetachedProcessAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}
