package main

import (
	"C"
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	HWND_BROADCAST = 0xffff
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	wSleep                  = kernel32.NewProc("Sleep")
	wOpenFileMappingW       = kernel32.NewProc("OpenFileMappingW")
	wMapViewOfFile          = kernel32.NewProc("MapViewOfFile")
	wCloseHandle            = kernel32.NewProc("CloseHandle")
	wUnmapViewOfFile        = kernel32.NewProc("UnmapViewOfFile")
	wOpenEvent              = kernel32.NewProc("OpenEventW")
	wWaitForSingleObject    = kernel32.NewProc("WaitForSingleObject")
	wRegisterWindowMessageA = kernel32.NewProc("RegisterWindowMessageA")
	wSendNotifyMessage      = kernel32.NewProc("SendNotifyMessage")
)

func sleep(timeout int) error {
	_, _, err := wSleep.Call(uintptr(timeout))

	if err != nil {
		errMsg := fmt.Sprintf("Timeout failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func openFileMapping(lpName string) (uintptr, error) {
	dwDesiredAccess := syscall.FILE_MAP_READ

	hMemMapFile, _, err := wOpenFileMappingW.Call(
		uintptr(dwDesiredAccess), // DWORD
		0, // BOOL
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpName))), // LPCTSTR
	)

	if hMemMapFile == 0 {
		errMsg := fmt.Sprintf("OpenFileMapping failed (%s)", err)
		return hMemMapFile, errors.New(errMsg)
	}

	return hMemMapFile, nil
}

func mapViewOfFile(hMemMapFile uintptr, dwNumberOfBytesToMap int) (uintptr, error) {
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

func closeHandle(handle uintptr) error {
	result, _, err := wCloseHandle.Call(handle)

	if result == 0 {
		errMsg := fmt.Sprintf("CloseHandle failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func unmapViewOfFile(lpBaseAddress uintptr) error {
	result, _, err := wUnmapViewOfFile.Call(lpBaseAddress)

	if result == 0 {
		errMsg := fmt.Sprintf("UnmapViewOfFile failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func openEvent(lpName string) (uintptr, error) {
	dwDesiredAccess := syscall.SYNCHRONIZE
	bInheritHandle := 0

	hDataValidEvent, _, err := wOpenEvent.Call(
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

func waitForSingleObject(hDataValidEvent uintptr, timeOut int) error {
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

func registerWindowMessageA(lpString string) (uint, error) {
	msgID, _, err := wRegisterWindowMessageA.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpString))), // LPCTSTR
	)

	if msgID == 0 {
		errMsg := fmt.Sprintf("registerWindowMessageA failed (%s)", err)
		return 0, errors.New(errMsg)
	}

	return uint(msgID), nil
}

func sendNotifyMessage(msgID uint, msg uint32, wParam uint32) error {
	hWnd := HWND_BROADCAST

	result, _, err := wSendNotifyMessage.Call(
		uintptr(hWnd), // HWND
		uintptr(msg), // UINT
		uintptr(wParam), // WPARAM
		0, // LPARAM
	)

	if result == 0 {
		errMsg := fmt.Sprintf("sendNotifyMessage failed (%s)", err)
		return errors.New(errMsg)
	}

	return nil
}

func MAKELONG(lo, hi uint16) uint32 {
	return uint32(uint32(lo) | ((uint32(hi)) << 16))
}
