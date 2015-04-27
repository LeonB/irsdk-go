// -build windows
package utils

import (
	"errors"
	"fmt"
	"net/rpc"
	"os/exec"

	"github.com/kevinwallace/coprocess"
)

type CWrapper struct {
	client *rpc.Client
}

// Syscalls

func (sw *CWrapper) OpenFileMapping(lpName string) (uintptr, error) {
	var handle uintptr
	args := &OpenFileMappingArgs{lpName}

	err := sw.client.Call("Commands.OpenFileMapping", args, &handle)
	return handle, err
}

func (sw *CWrapper) MapViewOfFile(hMemMapFile uintptr, dwNumberOfBytesToMap int) (uintptr, error) {
	var startAddress uintptr
	args := &MapViewOfFileArgs{hMemMapFile, dwNumberOfBytesToMap}

	err := sw.client.Call("Commands.MapViewOfFile", args, &startAddress)
	return startAddress, err
}

func (sw *CWrapper) CloseHandle(handle uintptr) error {
	return nil
}

func (sw *CWrapper) UnmapViewOfFile(lpBaseAddress uintptr) error {
	return nil
}

func (sw *CWrapper) OpenEvent(lpName string) (uintptr, error) {
	var handle uintptr
	args := &OpenEventArgs{lpName}

	err := sw.client.Call("Commands.OpenEvent", args, &handle)
	return handle, err
}

func (sw *CWrapper) WaitForSingleObject(hDataValidEvent uintptr, timeOut int) error {
	retVal := new(int)
	args := &WaitForSingleObjectArgs{hDataValidEvent, timeOut}

	return sw.client.Call("Commands.WaitForSingleObject", args, &retVal)
}

func (sw *CWrapper) RegisterWindowMessageA(lpString string) (uint, error) {
	return 0, nil
}

func (sw *CWrapper) SendNotifyMessage(msgID uint, wParam uint32, lParam uint32) error {
	return nil
}

// Pointer arithmetic

func (sw *CWrapper) ptrToHeader(sharedMemPtr uintptr) (*Header, error) {
	header := &Header{}
	args := &PtrToHeaderArgs{sharedMemPtr}

	err := sw.client.Call("Commands.PtrToHeader", args, header)
	return header, err
}

func (sw *CWrapper) ptrToSharedMem(sharedMemPtr uintptr) ([]byte, error) {
	sharedMem := new([]byte)
	args := &PtrToSharedMemArgs{sharedMemPtr}

	err := sw.client.Call("Commands.PtrToSharedMem", args, sharedMem)
	return *sharedMem, err
}

func (sw *CWrapper) ptrToVarHeader(varHeaderPtr uintptr) (*VarHeader, error) {
	varHeader := &VarHeader{}
	args := &PtrToVarHeaderArgs{varHeaderPtr}

	fmt.Println("1")
	err := sw.client.Call("Commands.PtrToVarHeader", args, varHeader)
	return varHeader, err
}

func (sw *CWrapper) getMemory(sharedMemPtr uintptr, startByte int, endByte int) ([]byte, error) {
	// bufLen := endByte - startByte
	data := new([]byte)

	// bytes := new([]byte)
	args := &GetMemoryArgs{sharedMemPtr, startByte, endByte}

	err := sw.client.Call("Commands.GetMemory", args, data)
	return *data, err
}

func NewCWrapper() (*CWrapper, error) {
	cmd := exec.Command("/opt/iracing/bin/wine", "--bottle", "default", "ir-syscalls-rpc.exe")
	client, err := coprocess.NewClient(cmd)
	if err != nil {
		return nil, err
	}

	args := new(string)
	*args = "ping"
	ret := new(bool)
	err = client.Call("Commands.Ping", args, ret)
	if err != nil {
		msg := fmt.Sprintf("Failed to execute rpc client (%v), make sure iRacing and ir-syscalls-rpc.exe are installed", err)
		err = errors.New(msg)
		return nil, err
	}

	sw := &CWrapper{client}
	return sw, nil
}
