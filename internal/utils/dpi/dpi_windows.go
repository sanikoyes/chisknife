//go:build windows
// +build windows

package dpi

import (
	"syscall"
)

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	shcore                     = syscall.NewLazyDLL("shcore.dll")
	procSetProcessDPIAware     = user32.NewProc("SetProcessDPIAware")
	procSetProcessDpiAwareness = shcore.NewProc("SetProcessDpiAwareness")
)

const (
	PROCESS_DPI_UNAWARE           = 0
	PROCESS_SYSTEM_DPI_AWARE      = 1
	PROCESS_PER_MONITOR_DPI_AWARE = 2
)

func init() {
	// 尝试使用 Windows 8.1+ 的 Per-Monitor DPI Aware
	if err := setProcessDpiAwareness(PROCESS_SYSTEM_DPI_AWARE); err != nil {
		// 如果失败，回退到 Windows Vista+ 的 System DPI Aware
		procSetProcessDPIAware.Call()
	}
}

func setProcessDpiAwareness(value int) error {
	ret, _, _ := procSetProcessDpiAwareness.Call(uintptr(value))
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}
