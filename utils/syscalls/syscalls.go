// +build windows

package syscalls

/*
// for timeBeginPeriod()
#pragma comment(lib, "Winmm")
// for RegisterWindowMessageA() and SendMessage()
#pragma comment(lib, "User32")
*/
import "C"
import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	HWND_BROADCAST = HWND(0xffff)
)

type (
	HWND   HANDLE
	HANDLE uintptr
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")

	wSleep                  = kernel32.NewProc("Sleep")
	wOpenFileMappingW       = kernel32.NewProc("OpenFileMappingW")
	wMapViewOfFile          = kernel32.NewProc("MapViewOfFile")
	wCloseHandle            = kernel32.NewProc("CloseHandle")
	wUnmapViewOfFile        = kernel32.NewProc("UnmapViewOfFile")
	wOpenEventW             = kernel32.NewProc("OpenEventW")
	wWaitForSingleObject    = kernel32.NewProc("WaitForSingleObject")
	wRegisterWindowMessageA = user32.NewProc("RegisterWindowMessageA")
	wSendNotifyMessageA     = user32.NewProc("SendNotifyMessageA")
)

func Sleep(timeout time.Duration) error {
	_, _, err := wSleep.Call(uintptr(timeout / time.Millisecond))

	if err != nil {
		errMsg := fmt.Sprintf("Timeout failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func OpenFileMapping(lpName string) (uintptr, error) {
	dwDesiredAccess := syscall.FILE_MAP_READ

	// Work around go bug
	// Without Println() wOpenFileMappingW.Call) fails
	fmt.Print("")

	hMemMapFile, _, err := wOpenFileMappingW.Call(
		uintptr(dwDesiredAccess), // DWORD
		0, // BOOL
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpName))), // LPCTSTR
	)

	if hMemMapFile == 0 {
		errMsg := fmt.Sprintf("OpenFileMappingW failed (%s)", err)
		return hMemMapFile, errors.New(errMsg)
	}

	return hMemMapFile, nil
}

func MapViewOfFile(hMemMapFile uintptr, dwNumberOfBytesToMap int) (uintptr, error) {
	dwDesiredAccess := syscall.FILE_MAP_READ
	dwFileOffsetHigh := 0
	dwFileOffsetLow := 0

	sharedMemPtr, _, err := wMapViewOfFile.Call(
		hMemMapFile,
		uintptr(dwDesiredAccess),      // DWORD
		uintptr(dwFileOffsetHigh),     // DWORD
		uintptr(dwFileOffsetLow),      // DWORD
		uintptr(dwNumberOfBytesToMap), // SIZE_T
	)

	if sharedMemPtr == 0 {
		errMsg := fmt.Sprintf("MapViewOfFile failed (%s)", err)
		return hMemMapFile, errors.New(errMsg)
	}

	return sharedMemPtr, nil
}

func CloseHandle(handle uintptr) error {
	result, _, err := wCloseHandle.Call(handle)

	if result == 0 {
		errMsg := fmt.Sprintf("CloseHandle failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func UnmapViewOfFile(lpBaseAddress uintptr) error {
	result, _, err := wUnmapViewOfFile.Call(lpBaseAddress)

	if result == 0 {
		errMsg := fmt.Sprintf("UnmapViewOfFile failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func OpenEvent(lpName string) (uintptr, error) {
	dwDesiredAccess := syscall.SYNCHRONIZE
	bInheritHandle := 0

	hDataValidEvent, _, err := wOpenEventW.Call(
		uintptr(dwDesiredAccess),                                  // DWORD
		uintptr(bInheritHandle),                                   // BOOL
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpName))), // LPCTSTR
	)

	if hDataValidEvent == 0 {
		errMsg := fmt.Sprintf("OpenEvent failed (%s)", err)
		return hDataValidEvent, errors.New(errMsg)
	}

	return hDataValidEvent, nil
}

func WaitForSingleObject(hDataValidEvent uintptr, timeOut int) error {
	dwMilliseconds := timeOut

	result, _, err := wWaitForSingleObject.Call(
		hDataValidEvent,         // HANDLE
		uintptr(dwMilliseconds), // DWORD
	)

	if result != 0 {
		errMsg := fmt.Sprintf("WaitForSingleObject failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func RegisterWindowMessageA(lpString string) (uint, error) {
	msgID, _, err := wRegisterWindowMessageA.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpString))), // LPCTSTR
	)

	if msgID == 0 {
		errMsg := fmt.Sprintf("registerWindowMessageA failed (%s)", err)
		return 0, errors.New(errMsg)
	}

	return uint(msgID), nil
}

func SendNotifyMessage(msgID uint, wParam uint32, lParam uint32) error {
	hWnd := HWND_BROADCAST

	result, _, err := wSendNotifyMessageA.Call(
		uintptr(hWnd),   // HWND
		uintptr(msgID),  // UINT
		uintptr(wParam), // WPARAM
		uintptr(lParam), // LPARAM
	)

	fmt.Println(err)
	if result == 0 {
		errMsg := fmt.Sprintf("sendNotifyMessage failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func Now() time.Time {
	t := &syscall.Timeval{}
	syscall.Gettimeofday(t)
	sec, _ := t.Unix()
	nSec := t.Nano()

	return time.Unix(sec, nSec)
}
