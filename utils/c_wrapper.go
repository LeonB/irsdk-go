package utils

import "unsafe"

const (
)

func (cw *CWrapper) getHeader() (*Header, error) {
	return (*Header)(cw.sharedMemPtr), nil
}

func (cw *CWrapper) getVarHeaderEntry(index int) (*VarHeader, error) {
	varHeader := &VarHeader{}
	header := cw.header
	sharedMemPtr := cw.sharedMemPtr

	varHeaderOffset := uintptr(header.VarHeaderOffset)
	varHeaderSize := uintptr(unsafe.Sizeof(*varHeader))
	i := uintptr(index)
	totalOffset := varHeaderOffset + (varHeaderSize * i)
	varHeaderPtr := uintptr(sharedMemPtr) + totalOffset

	return (*VarHeader)(unsafe.Pointer(varHeaderPtr)), nil
}
