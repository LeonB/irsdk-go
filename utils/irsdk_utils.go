package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"time"
	"unsafe"
)

const (
	MEMMAPFILENAME     = "Local\\IRSDKMemMapFileName"
	BROADCASTMSGNAME   = "IRSDK_BROADCASTMSG"
	DATAVALIDEVENTNAME = "Local\\IRSDKDataValidEvent"
	INT_MAX            = 2147483647
	MEMMAPFILESIZE     = 780 * 1024

	MAX_BUFS   = 4
	MAX_STRING = 32
	// descriptions can be longer than max_string!
	MAX_DESC = 64

	TIMEOUT = time.Duration(30) // timeout after 30 seconds with no communication
)

var (
	ErrInitialize     = errors.New("Failed to initialize")
	ErrDataChanged    = errors.New("Data changed out from under us")
	ErrDisconnected   = errors.New("We probably disconnected")
	ErrNothingChanged = errors.New("Nothing changed this tick")
)

type Irsdk struct {
	hDataValidEvent uintptr
	hMemMapFile     uintptr
	sharedMemPtr    uintptr
	sharedMem       []byte

	header        *Header
	isInitialized bool
	lastValidTime time.Time
	lastTickCount int

	rpcCmd *exec.Cmd
}

func (ir *Irsdk) Startup() error {
	var err error

	if ir.hMemMapFile == 0 {
		ir.hMemMapFile, err = openFileMapping(MEMMAPFILENAME)
		if err != nil {
			return err
		}
		ir.lastTickCount = INT_MAX
	}

	if ir.hMemMapFile != 0 {
		if len(ir.sharedMem) == 0 {
			ir.sharedMemPtr, err = mapViewOfFile(ir.hMemMapFile, MEMMAPFILESIZE)
			if err != nil {
				return err
			}

			ir.header = (*Header)(unsafe.Pointer(ir.sharedMemPtr))
			ir.sharedMem = (*[MEMMAPFILESIZE]byte)(unsafe.Pointer(ir.sharedMemPtr))[:]
			ir.lastTickCount = INT_MAX
		}

		if len(ir.sharedMem) != 0 {
			if ir.hDataValidEvent == 0 {
				ir.hDataValidEvent, err = openEvent(DATAVALIDEVENTNAME)
				if err != nil {
					return err
				}

				ir.lastTickCount = INT_MAX
			}

			if ir.hDataValidEvent != 0 {
				ir.isInitialized = true
				return nil
			}
			//else printf("Error opening event: %d\n", GetLastError());
		}
		//else printf("Error mapping file: %d\n", GetLastError());
	}
	//else printf("Error opening file: %d\n", GetLastError()); `

	ir.isInitialized = false
	return ErrInitialize
}

func (ir *Irsdk) Shutdown() {
	if ir.hDataValidEvent != 0 {
		closeHandle(ir.hDataValidEvent)

		if len(ir.sharedMem) != 0 {
			ir.sharedMemPtr = uintptr(unsafe.Pointer(&ir.sharedMem))
			unmapViewOfFile(ir.sharedMemPtr)

			if ir.hMemMapFile != 0 {
				closeHandle(ir.hMemMapFile)

				ir.hDataValidEvent = 0
				ir.sharedMem = nil
				ir.header = nil
				ir.hMemMapFile = 0

				ir.isInitialized = false
				ir.lastTickCount = INT_MAX
			}
		}
	}
}

func (ir *Irsdk) GetNewData() ([]byte, error) {
	if !ir.isInitialized {
		err := ir.Startup()
		if err != nil {
			return nil, err
		}
	}

	// if sim is not active, then no new data
	if (ir.header.Status & StatusConnected) == 0 {
		ir.lastTickCount = INT_MAX
		return nil, nil
	}

	latest := 0
	for i := 0; i < int(ir.header.NumBuf); i++ {
		if ir.header.VarBuf[latest].TickCount < ir.header.VarBuf[i].TickCount {
			latest = i
		}
	}

	// if newer than last recieved, than report new data
	if ir.lastTickCount < int(ir.header.VarBuf[latest].TickCount) {

		for count := 0; count < 2; count++ {
			curTickCount := int(ir.header.VarBuf[latest].TickCount)
			bufLen := int(ir.header.BufLen)
			startByte := int(ir.header.VarBuf[latest].BufOffset)
			endByte := startByte + bufLen

			// Copy data
			data := make([]byte, bufLen)
			copy(data, ir.sharedMem[startByte:endByte])

			if curTickCount == int(ir.header.VarBuf[latest].TickCount) {
				ir.lastTickCount = curTickCount
				ir.lastValidTime = now()
				return data, nil
			}
		}
		// if here, the data changed out from under us.
		return nil, ErrDataChanged
	} else if ir.lastTickCount > int(ir.header.VarBuf[latest].TickCount) {
		// if older than last recieved, than reset, we probably disconnected
		ir.lastTickCount = int(ir.header.VarBuf[latest].TickCount)
		return nil, ErrDisconnected
	}

	// else the same, and nothing changed this tick
	return nil, ErrNothingChanged
}

func (ir *Irsdk) WaitForDataReady(timeOut time.Duration) ([]byte, error) {
	var data []byte
	var err error

	if !ir.isInitialized {
		err = ir.Startup()

		if err != nil {
			// sleep if error
			if timeOut > 0 {
				sleep(timeOut)
			}

			return nil, nil
		}
	}

	// just to be sure, check before we sleep
	data, err = ir.GetNewData()
	if data != nil {
		return data, err
	}

	// sleep till signaled
	waitForSingleObject(ir.hDataValidEvent, int(timeOut/time.Millisecond))

	// we woke up, so check for data
	data, err = ir.GetNewData()
	return data, err
}

func (ir *Irsdk) IsConnected() bool {
	if ir.isInitialized {
		elapsed := now().Sub(ir.lastValidTime)
		if (ir.header.Status & StatusConnected) > 0 && (elapsed < TIMEOUT) {
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

func (ir *Irsdk) GetHeader() (*Header, error) {
	return nil, nil
}

func (ir *Irsdk) GetVarHeader() (*VarHeader, error) {
	return nil, nil
}

func (ir *Irsdk) GetVarHeaderEntry(index int) *VarHeader {
	if ir.isInitialized {
		if index >= 0 && index < (int)(ir.header.NumVars) {
			varHeader := &VarHeader{}
			pSharedMemPtr := uintptr(unsafe.Pointer(&ir.sharedMem[0]))
			varHeaderOffset := uintptr(ir.header.VarHeaderOffset)
			varHeaderSize := uintptr(unsafe.Sizeof(*varHeader))
			i := uintptr(index)
			totalOffset := varHeaderOffset + (varHeaderSize * i)
			varHeaderPtr := pSharedMemPtr + totalOffset

			varHeader = (*VarHeader)(unsafe.Pointer(varHeaderPtr))

			return varHeader
		}
	}
	return nil
}

// Note: this is a linear search, so cache the results
func (ir *Irsdk) VarNameToIndex(name string) int {
	var pVar *VarHeader

	if name != "" {
		numVars := int(ir.header.NumVars)
		for index := 0; index <= numVars; index++ {
			pVar = ir.GetVarHeaderEntry(index)
			pVarName := CToGoString(pVar.Name[:])
			if pVar != nil && pVarName == name {
				return index
			}
		}
	}

	return -1
}

func (ir *Irsdk) VarNameToOffset(name string) int {
	var pVar *VarHeader

	if name != "" {
		numVars := int(ir.header.NumVars)
		for index := 0; index <= numVars; index++ {
			pVar = ir.GetVarHeaderEntry(index)
			pVarName := CToGoString(pVar.Name[:])
			if pVar != nil && pVarName == name {
				return int(pVar.Offset)
			}
		}
	}

	return -1
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
		err := sendNotifyMessage(msgID, wParam, lParam)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ir *Irsdk) PadCarNum(num int, zero int) int {
	retVal := num
	numPlace := 1
	if num > 99 {
		numPlace = 3
	} else if num > 9 {
		numPlace = 2
	}
	if zero != 0 {
		numPlace += zero
		retVal = num + 1000*numPlace
	}

	return retVal
}

// Custom functions

func (ir *Irsdk) GetRpcCmd() (*exec.Cmd, error) {
	return nil, nil
}

func (ir *Irsdk) GetNumVars() int {
	return int(ir.header.NumVars)
}

func (ir *Irsdk) GetBroadcastMsgID() (uint, error) {
	return registerWindowMessageA(BROADCASTMSGNAME)
}

func (ir *Irsdk) GetSharedMem() []byte {
	return ir.sharedMem
}

func CToGoString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}
