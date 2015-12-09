// +build windows

package irsdk

import (
	"time"
	"unsafe"

	syscalls "github.com/leonb/irsdk-go/syscalls"
)

type CWrapper struct {
	sharedMemPtr    unsafe.Pointer
	sharedMem       []byte
	header          *Header
	hDataValidEvent uintptr

	hMemMapFile uintptr
}

func (cw *CWrapper) startup() error {
	var err error

	if cw.hMemMapFile == 0 {
		cw.hMemMapFile, err = cw.OpenFileMapping(MEMMAPFILENAME)
		if err != nil {
			return err
		}
	}

	if len(cw.sharedMem) == 0 {
		sharedMemPtr, err := cw.MapViewOfFile(cw.hMemMapFile, MEMMAPFILESIZE)
		cw.sharedMemPtr = unsafe.Pointer(sharedMemPtr)
		if err != nil {
			return err
		}
	}

	if cw.header == nil {
		cw.header = (*Header)(cw.sharedMemPtr)
	}

	if cw.sharedMem == nil {
		cw.sharedMem = (*[MEMMAPFILESIZE]byte)(cw.sharedMemPtr)[:]
	}

	if cw.hDataValidEvent == 0 {
		cw.hDataValidEvent, err = cw.OpenEvent(DATAVALIDEVENTNAME)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cw *CWrapper) shutdown() error {
	if cw.hDataValidEvent != 0 {
		cw.CloseHandle(cw.hDataValidEvent)
	}

	if cw.sharedMemPtr != nil {
		cw.UnmapViewOfFile(uintptr(cw.sharedMemPtr))
	}

	if cw.hMemMapFile != 0 {
		cw.CloseHandle(cw.hMemMapFile)
	}

	// Clean windows specific vars
	cw.hDataValidEvent = 0
	cw.hMemMapFile = 0

	// Clean global vars
	cw.sharedMemPtr = nil
	cw.sharedMem = nil
	cw.header = nil

	return nil
}

func (cw *CWrapper) WaitForDataChange(timeout time.Duration) error {
	return cw.WaitForSingleObject(cw.hDataValidEvent, int(timeout/time.Millisecond))
}

// Syscalls

func (cw *CWrapper) OpenFileMapping(lpName string) (uintptr, error) {
	return syscalls.OpenFileMapping(lpName)
}

func (cw *CWrapper) MapViewOfFile(hMemMapFile uintptr, dwNumberOfBytesToMap int) (uintptr, error) {
	return syscalls.MapViewOfFile(hMemMapFile, dwNumberOfBytesToMap)
}

func (cw *CWrapper) CloseHandle(handle uintptr) error {
	return syscalls.CloseHandle(handle)
}

func (cw *CWrapper) UnmapViewOfFile(lpBaseAddress uintptr) error {
	return syscalls.UnmapViewOfFile(lpBaseAddress)
}

func (cw *CWrapper) OpenEvent(lpName string) (uintptr, error) {
	return syscalls.OpenEvent(lpName)
}

func (cw *CWrapper) WaitForSingleObject(hDataValidEvent uintptr, timeOut int) error {
	return syscalls.WaitForSingleObject(hDataValidEvent, timeOut)
}

func (cw *CWrapper) RegisterWindowMessageA(lpString string) (uint, error) {
	return syscalls.RegisterWindowMessageA(lpString)
}

func (cw *CWrapper) RegisterWindowMessageW(lpString string) (uint, error) {
	return syscalls.RegisterWindowMessageW(lpString)
}

func (cw *CWrapper) SendNotifyMessageW(msgID uint, wParam uint32, lParam uint32) error {
	return syscalls.SendNotifyMessageW(msgID, wParam, lParam)
}

func NewCWrapper() (*CWrapper, error) {
	sw := &CWrapper{}
	return sw, nil
}
