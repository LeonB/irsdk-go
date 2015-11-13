package utils

import "unsafe"

const ()

func (cw *CWrapper) getHeader() (*Header, error) {
	return (*Header)(cw.sharedMemPtr), nil
}

func (cw *CWrapper) getVarHeaderEntry(index int) (*VarHeader, error) {
	varHeaderRaw := &VarHeaderRaw{}
	header := cw.header
	sharedMemPtr := cw.sharedMemPtr

	varHeaderOffset := uintptr(header.VarHeaderOffset)
	varHeaderSize := uintptr(unsafe.Sizeof(*varHeaderRaw))
	i := uintptr(index)
	totalOffset := varHeaderOffset + (varHeaderSize * i)
	varHeaderPtr := uintptr(sharedMemPtr) + totalOffset

	varHeaderRaw = (*VarHeaderRaw)(unsafe.Pointer(varHeaderPtr))
	varHeader := &VarHeader{
		Type:   varHeaderRaw.Type,
		Offset: varHeaderRaw.Offset,
		Count:  varHeaderRaw.Count,

		Name: CToGoString(varHeaderRaw.Name[:]),
		Desc: CToGoString(varHeaderRaw.Desc[:]),
		Unit: CToGoString(varHeaderRaw.Unit[:]),
	}
	return varHeader, nil
}
