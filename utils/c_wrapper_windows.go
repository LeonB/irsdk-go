// +build windows

package utils

import (
	"unsafe"

	syscalls "github.com/leonb/irsdk-go/utils/syscalls"
)

type CWrapper struct{}

// Syscalls

func (sw *CWrapper) OpenFileMapping(lpName string) (uintptr, error) {
	return syscalls.OpenFileMapping(lpName)
}

func (sw *CWrapper) MapViewOfFile(hMemMapFile uintptr, dwNumberOfBytesToMap int) (uintptr, error) {
	return syscalls.MapViewOfFile(hMemMapFile, dwNumberOfBytesToMap)
}

func (sw *CWrapper) CloseHandle(handle uintptr) error {
	return syscalls.CloseHandle(handle)
}

func (sw *CWrapper) UnmapViewOfFile(lpBaseAddress uintptr) error {
	return syscalls.UnmapViewOfFile(lpBaseAddress)
}

func (sw *CWrapper) OpenEvent(lpName string) (uintptr, error) {
	return syscalls.OpenEvent(lpName)
}

func (sw *CWrapper) WaitForSingleObject(hDataValidEvent uintptr, timeOut int) error {
	return syscalls.WaitForSingleObject(hDataValidEvent, timeOut)
}

func (sw *CWrapper) RegisterWindowMessageA(lpString string) (uint, error) {
	return syscalls.RegisterWindowMessageA(lpString)
}

func (sw *CWrapper) SendNotifyMessage(msgID uint, wParam uint32, lParam uint32) error {
	return syscalls.SendNotifyMessage(msgID, wParam, lParam)
}

// Pointer arithmetic

func (sw *CWrapper) ptrToHeader(sharedMemPtr uintptr) (*Header, error) {
	return (*Header)(unsafe.Pointer(sharedMemPtr)), nil
}

func (sw *CWrapper) ptrToSharedMem(sharedMemPtr uintptr) ([]byte, error) {
	return (*[MEMMAPFILESIZE]byte)(unsafe.Pointer(sharedMemPtr))[:], nil
}

func (sw *CWrapper) ptrToVarHeader(varHeaderPtr uintptr) (*VarHeader, error) {
	return (*VarHeader)(unsafe.Pointer(varHeaderPtr)), nil
}

func (sw *CWrapper) getMemory(sharedMemPtr uintptr, startByte int, endByte int) ([]byte, error) {
	sharedMem := (*[MEMMAPFILESIZE]byte)(unsafe.Pointer(sharedMemPtr))[:]
	bufLen := endByte - startByte
	data := make([]byte, bufLen)
	copy(data, sharedMem[startByte:endByte])
	return data, nil
}

func NewCWrapper() (*CWrapper, error) {
	sw := &CWrapper{}
	return sw, nil
}
