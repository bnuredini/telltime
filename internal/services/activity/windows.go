//go:build windows
// +build windows

package activity

import (
	"database/sql"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/bnuredini/telltime/internal/conf"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")

	procSetWinEventHook  = user32.NewProc("SetWinEventHook")
	procUnhookWinEvent   = user32.NewProc("UnhookWinEvent")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")

	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procGetModuleBaseNameW       = psapi.NewProc("GetModuleBaseNameW")
)

const (
	EVENT_SYSTEM_FOREGROUND = 0x0003
	WINEVENT_OUTOFCONTEXT   = 0x0000

	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct {
		X int32
		Y int32
	}
}

func initWindows(db *sql.DB, config *conf.Config) {
	hookCallback := syscall.NewCallback(winEventProc)

	hHook, _, err := procSetWinEventHook.Call(
		uintptr(EVENT_SYSTEM_FOREGROUND),
		uintptr(EVENT_SYSTEM_FOREGROUND),
		0,
		hookCallback,
		0,
		0,
		uintptr(WINEVENT_OUTOFCONTEXT),
	)

	if hHook == 0 {
		fmt.Println("Failed to set hook:", err)
		return
	}

	fmt.Println("Listening for active window changes...")

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) == -1 {
			break
		}

		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	procUnhookWinEvent.Call(hHook)
}

func winEventProc(
	hWinEventHook uintptr,
	event uint32,
	hwnd syscall.Handle,
	idObject uint32,
	idChild uint32,
	dwEventThread uint32,
	dwmsEventTime uint32,
) uintptr {

	if hwnd == 0 {
		return 0
	}

	var pid uint32
	procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return 0
	}

	hProcess, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		0,
		uintptr(pid),
	)
	if hProcess == 0 {
		return 0
	}
	defer procCloseHandle.Call(hProcess)

	exeName := make([]uint16, 260)
	ret, _, _ := procGetModuleBaseNameW.Call(
		hProcess,
		0,
		uintptr(unsafe.Pointer(&exeName[0])),
		uintptr(len(exeName)),
	)
	if ret == 0 {
		return 0
	}

	processName := syscall.UTF16ToString(exeName)

	fmt.Printf("Active application changed: %s\n", processName)

	return 0
}
