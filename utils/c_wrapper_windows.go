// +build windows

package utils

import (
	"unsafe"

	syscalls "github.com/leonb/irsdk-go/utils/syscalls"
)

type CWrapper struct {
	hDataValidEvent uintptr
	hMemMapFile     uintptr
	sharedMemPtr    uintptr
	sharedMem       []byte
}

func (cw *CWrapper) startup() error {
	if cw.hDataValidEvent == 0 {
		cw.hDataValidEvent, err = cw.c.OpenEvent(DATAVALIDEVENTNAME)
		if err != nil {
			return err
		}
	}
}

func (cw *CWrapper) WaitForDataChange(timeOut time.Duration) error {
	return cw.WaitForSingleObject(cw.hDataValidEvent, int(timeOut/time.Millisecond))
}

func (ir *Irsdk) IsConnected() bool {
	if ir.isInitialized {
		elapsed := time.Now().Sub(ir.lastValidTime)
		if (ir.header.Status&StatusConnected) > 0 && (elapsed < TIMEOUT) {
			return true
		}
	}

	return false
}

func (ir *Irsdk) GetSessionInfoStr() []byte {
	if ir.isInitialized {
		startByte := ir.header.SessionInfoOffset
		length := ir.header.SessionInfoLen
		return ir.sharedMem[startByte:length]
	}
	return nil
}

func (ir *Irsdk) GetVarHeaderEntry(index int) (*VarHeader, error) {
	if ir.isInitialized {
		if index >= 0 && index < (int)(ir.header.NumVars) {
			return ir.c.getVarHeaderEntry(index)
		}
	}
	return nil, nil
}

// Note: this is a linear search, so cache the results
func (ir *Irsdk) VarNameToIndex(name string) (int, error) {
	if name != "" {
		numVars := int(ir.header.NumVars)
		for index := 0; index <= numVars; index++ {
			pVar, err := ir.GetVarHeaderEntry(index)
			if err != nil {
				return -1, err
			}
			pVarName := CToGoString(pVar.Name[:])
			if pVar != nil && pVarName == name {
				return index, nil
			}
		}
	}

	return -1, nil
}

func (ir *Irsdk) VarNameToOffset(name string) (int, error) {
	if name != "" {
		numVars := int(ir.header.NumVars)
		for index := 0; index <= numVars; index++ {
			pVar, err := ir.GetVarHeaderEntry(index)
			if err != nil {
				return -1, err
			}
			pVarName := CToGoString(pVar.Name[:])
			if pVar != nil && pVarName == name {
				return int(pVar.Offset), nil
			}
		}
	}

	return -1, nil
}

func (ir *Irsdk) BroadcastMsg(msg BroadcastMsg, var1 uint16, var2 uint16, var3 uint16) error {
	msgID, _ := ir.GetBroadcastMsgID()

	wParam := MAKELONG(uint16(msg), var1)
	lParam := MAKELONG(var2, var3)

	fmt.Println("msgID:", msgID)
	fmt.Println("msg:", msg)
	fmt.Println("var1:", var1)
	fmt.Println("var2:", var2)
	fmt.Println("var3:", var3)
	fmt.Println("wParam", wParam)
	fmt.Println("lParam", lParam)

	if msgID > 0 && msg >= 0 && msg < BroadcastLast {
		err := ir.c.SendNotifyMessage(msgID, wParam, lParam)
		if err != nil {
			return err
		}
	}

	return nil
}

// Custom functions

// func (ir *Irsdk) GetRpcCmd() (*exec.Cmd, error) {
// 	return nil, nil
// }

func (ir *Irsdk) GetNumVars() int {
	return int(ir.header.NumVars)
}

func (ir *Irsdk) GetBroadcastMsgID() (uint, error) {
	return ir.c.RegisterWindowMessageA(BROADCASTMSGNAME)
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

func (cw *CWrapper) SendNotifyMessage(msgID uint, wParam uint32, lParam uint32) error {
	return syscalls.SendNotifyMessage(msgID, wParam, lParam)
}

func NewCWrapper() (*CWrapper, error) {
	sw := &CWrapper{}
	return sw, nil
}
