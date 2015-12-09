// +build windows

package irsdk

import (
	"unsafe"

	syscalls "github.com/leonb/irsdk-go/syscalls"
)

type RpcCommands struct{}

func (c *RpcCommands) Ping(args *string, ret *bool) error {
	*ret = true
	return nil
}

func (c *RpcCommands) OpenFileMapping(args *OpenFileMappingArgs, handle *uintptr) error {
	v, err := syscalls.OpenFileMapping(args.LpName)
	*handle = uintptr(v)
	return err
}

func (c *RpcCommands) MapViewOfFile(args *MapViewOfFileArgs, startAddress *uintptr) error {
	v, err := syscalls.MapViewOfFile(args.HMemMapFile, args.DwNumberOfBytesToMap)
	*startAddress = v
	return err
}

func (c *RpcCommands) OpenEvent(args *OpenEventArgs, handle *uintptr) error {
	v, err := syscalls.OpenEvent(args.LpName)
	*handle = uintptr(v)
	return err
}

func (c *RpcCommands) WaitForSingleObject(args *WaitForSingleObjectArgs, retVal *int) error {
	return syscalls.WaitForSingleObject(args.HDataValidEvent, args.TimeOut)
}

func (c *RpcCommands) PtrToHeader(args *PtrToHeaderArgs, header *Header) error {
	*header = *(*Header)(unsafe.Pointer(args.SharedMemPtr))
	return nil
}

func (c *RpcCommands) PtrToSharedMem(args *PtrToSharedMemArgs, sharedMem *[]byte) error {
	*sharedMem = (*[MEMMAPFILESIZE]byte)(unsafe.Pointer(args.SharedMemPtr))[:]
	return nil
}

func (c *RpcCommands) PtrToVarHeader(args *PtrToVarHeaderArgs, varHeader *VarHeader) error {
	*varHeader = *(*VarHeader)(unsafe.Pointer(uintptr(args.VarHeaderPtr)))
	return nil
}

func (c *RpcCommands) GetMemory(args *GetMemoryArgs, data *[]byte) error {
	sharedMem := (*[MEMMAPFILESIZE]byte)(unsafe.Pointer(args.SharedMemPtr))[:]
	*data = sharedMem[args.StartByte:args.EndByte]
	return nil
}
