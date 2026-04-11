//go:build windows

package platform

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

func ensureUserPathEntry(dir string) (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return false, fmt.Errorf(`open HKCU\\Environment: %w`, err)
	}
	defer key.Close()

	currentValue, _, err := key.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return false, fmt.Errorf(`read user PATH from HKCU\\Environment: %w`, err)
	}
	if pathListContainsDir(currentValue, dir, true) {
		return false, nil
	}

	newValue := strings.TrimSpace(currentValue)
	if newValue == "" {
		newValue = filepath.Clean(dir)
	} else {
		newValue = newValue + ";" + filepath.Clean(dir)
	}

	if err := key.SetExpandStringValue("Path", newValue); err != nil {
		return false, fmt.Errorf(`write user PATH to HKCU\\Environment: %w`, err)
	}
	_ = broadcastEnvironmentChange()
	return true, nil
}

func broadcastEnvironmentChange() error {
	const (
		hwndBroadcast   = 0xffff
		wmSettingChange = 0x001A
		smtoAbortIfHung = 0x0002
	)

	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("SendMessageTimeoutW")
	name, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return err
	}

	_, _, callErr := proc.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(name)),
		uintptr(smtoAbortIfHung),
		5000,
		0,
	)
	if callErr != syscall.Errno(0) {
		return callErr
	}
	return nil
}
