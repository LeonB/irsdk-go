package main

import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"encoding/binary"
)

const (
	IRSDK_MEMMAPFILENAME     = "Local\\IRSDKMemMapFileName"
	IRSDK_BROADCASTMSGNAME   = "IRSDK_BROADCASTMSG"
	IRSDK_DATAVALIDEVENTNAME = "Local\\IRSDKDataValidEvent"
	INT_MAX                  = 2147483647
	MEMMAPFILESIZE           = 780 * 1024

	IRSDK_MAX_BUFS   = 4
	IRSDK_MAX_STRING = 32
	// descriptions can be longer than max_string!
	IRSDK_MAX_DESC = 64

	TIMEOUT = time.Duration(30) // timeout after 30 seconds with no communication
)

// Local memory

var hDataValidEvent uintptr
var hMemMapFile uintptr

var pHeader *irsdk_header
var isInitialized bool
var lastValidTime time.Time
var timeout time.Duration
var pSharedMem []byte

// var sharedMemPtr uintptr
var lastTickCount = INT_MAX

func irsdk_startup() error {
	var err error

	if hMemMapFile == 0 {
		hMemMapFile, err = openFileMapping(IRSDK_MEMMAPFILENAME)
		if err != nil {
			fmt.Println(err)
			return err
		}
		lastTickCount = INT_MAX
	}

	if hMemMapFile != 0 {
		if len(pSharedMem) == 0 {
			sharedMemPtr, err := mapViewOfFile(hMemMapFile, MEMMAPFILESIZE)
			if err != nil {
				return err
			}

			pHeader = (*irsdk_header)(unsafe.Pointer(sharedMemPtr))
			pSharedMem = (*[1 << 30]byte)(unsafe.Pointer(sharedMemPtr))[:]
			lastTickCount = INT_MAX
		}

		if len(pSharedMem) != 0 {
			if hDataValidEvent == 0 {
				// hDataValidEvent = try.N("OpenEvent", SYNCHRONIZE, false, syscall.StringToUTF16Ptr(IRSDK_DATAVALIDEVENTNAME))
				hDataValidEvent, err = openEvent(IRSDK_DATAVALIDEVENTNAME)
				if err != nil {
					return err
				}

				lastTickCount = INT_MAX
			}

			if hDataValidEvent != 0 {
				isInitialized = true
				return nil
			}
			//else printf("Error opening event: %d\n", GetLastError());
		}
		//else printf("Error mapping file: %d\n", GetLastError());
	}
	//else printf("Error opening file: %d\n", GetLastError()); `

	isInitialized = false
	return errors.New("Failed to initialize")
}

func irsdk_shutdown() {
	if hDataValidEvent != 0 {
		closeHandle(hDataValidEvent)

		if len(pSharedMem) != 0 {
			sharedMemPtr := uintptr(unsafe.Pointer(&pSharedMem))
			unmapViewOfFile(sharedMemPtr)

			if hMemMapFile != 0 {
				closeHandle(hMemMapFile)

				hDataValidEvent = 0
				pSharedMem = nil
				pHeader = nil
				hMemMapFile = 0

				isInitialized = false
				lastTickCount = INT_MAX
			}
		}
	}
}

func irsdk_getNewData() ([]byte, error) {
	if !isInitialized {
		err := irsdk_startup()
		if err != nil {
			return nil, err
		}
	}

	// if sim is not active, then no new data
	if (int(pHeader.Status) & int(irsdk_stConnected)) == 0 {
		lastTickCount = INT_MAX
		return nil, nil
	}

	latest := 0
	for i := 0; i < int(pHeader.NumBuf); i++ {
		if pHeader.VarBuf[latest].TickCount < pHeader.VarBuf[i].TickCount {
			latest = i
		}
	}

	// if newer than last recieved, than report new data
	if lastTickCount < int(pHeader.VarBuf[latest].TickCount) {

		for count := 0; count < 2; count++ {
			curTickCount := int(pHeader.VarBuf[latest].TickCount)
			bufLen := int(pHeader.BufLen)
			startByte := int(pHeader.VarBuf[latest].BufOffset)
			endByte := startByte + bufLen

			// Copy data
			data := make([]byte, bufLen)
			copy(data, pSharedMem[startByte:endByte])
			// data := pSharedMem[startByte:endByte]

			if curTickCount == int(pHeader.VarBuf[latest].TickCount) {
				lastTickCount = curTickCount
				lastValidTime = time.Now()
				return data, nil
			}
		}
		// if here, the data changed out from under us.
		return nil, errors.New("Data changed out from under us")
	} else if lastTickCount > int(pHeader.VarBuf[latest].TickCount) {
		// if older than last recieved, than reset, we probably disconnected
		lastTickCount = int(pHeader.VarBuf[latest].TickCount)
		return nil, errors.New("We probably disconnected")
	}

	// else the same, and nothing changed this tick
	return nil, errors.New("Nothing changed this tick")
}

func irsdk_waitForDataReady(timeOut int) ([]byte, error) {
	var data []byte
	var err error

	if !isInitialized {
		err = irsdk_startup()

		if err != nil {
			// sleep if error
			// @TODO: fix this
			if timeOut > 0 {
				sleep(timeOut)
			}

			return nil, nil
		}
	}

	// just to be sure, check before we sleep
	data, err = irsdk_getNewData()
	if err == nil {
		return data, err
	}

	// sleep till signaled
	waitForSingleObject(hDataValidEvent, timeOut)

	// we woke up, so check for data
	data, err = irsdk_getNewData()
	if err != nil {
		return nil, err
	}

	return data, err
}
func irsdk_isConnected() bool {
	if isInitialized {
		elapsed := time.Now().Sub(lastValidTime)
		if (pHeader.Status&irsdk_stConnected) > 0 && (elapsed < timeout) {
			return true
		}
	}

	return false
}

// direct access to the data buffer
// // Warnign! This buffer is volitile so read it out fast!
// // Use the cached copy from irsdk_waitForDataReady() or irsdk_getNewData()
// instead
func irsdk_getData(index int) []byte {
	if isInitialized {
		endByte := int(pHeader.VarBuf[index].BufOffset)
		return pSharedMem[:endByte]
	}

	return nil
}

func irsdk_getSessionInfoStr() []byte {
	if isInitialized {
		return pSharedMem[pHeader.SessionInfoOffset:pHeader.SessionInfoLen]
	}
	return nil
}

func irsdk_getVarHeaderPtr() *irsdk_varHeader {
	if isInitialized {
		varHeaderOffset := int(pHeader.VarHeaderOffset)
		varHeader := &irsdk_varHeader{}
		varHeaderSize := int(unsafe.Sizeof(*varHeader))

		startByte := varHeaderOffset
		endByte := startByte + varHeaderSize

		// create a io.Reader
		b := bytes.NewBuffer(pSharedMem[startByte:endByte])
		// read []byte and convert it into irsdk_varHeader
		binary.Read(b, binary.LittleEndian, varHeader)

		return varHeader
	}
	return nil
}

func irsdk_getVarHeaderEntry(index int) *irsdk_varHeader {
	if isInitialized {
		if index >= 0 && index < (int)(pHeader.NumVars) {
			varHeaderOffset := int(pHeader.VarHeaderOffset)
			varHeader := &irsdk_varHeader{}
			varHeaderSize := int(unsafe.Sizeof(*varHeader))

			startByte := varHeaderOffset + (index * varHeaderSize)
			endByte := startByte + varHeaderSize

			// create a io.Reader
			b := bytes.NewBuffer(pSharedMem[startByte:endByte])
			// read []byte and convert it into irsdk_varHeader
			binary.Read(b, binary.LittleEndian, varHeader)

			return varHeader
		}
	}
	return nil
}

// Note: this is a linear search, so cache the results
func irsdk_varNameToIndex(name string) int {
	var pVar *irsdk_varHeader

	if name != "" {
		numVars := int(pHeader.NumVars)
		for index := 0; index <= numVars; index++ {
			pVar = irsdk_getVarHeaderEntry(index)
			pVarName := CToGoString(pVar.Name[:])
			if pVar != nil && pVarName == name {
				return index
			}
		}
	}

	return -1
}

func irsdk_varNameToOffset(name string) C.int {
	var pVar *irsdk_varHeader

	if name != "" {
		numVars := int(pHeader.NumVars)
		for index := 0; index <= numVars; index++ {
			pVar = irsdk_getVarHeaderEntry(index)
			pVarName := CToGoString(pVar.Name[:])
			if pVar != nil && pVarName == name {
				return pVar.Offset
			}
		}
	}

	return -1
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

func irsdk_getBroadcastMsgID() (uint, error) {
	return registerWindowMessageA(IRSDK_BROADCASTMSGNAME)
}

func irsdk_broadcastMsg(msg irsdk_BroadcastMsg, var1 uint16, var2 uint16, var3 uint16) {
	msgID, _ := irsdk_getBroadcastMsgID()

	wParam := MAKELONG(var2, var3)
	msg2 := MAKELONG(uint16(msg), var1)

	if msgID > 0 && msg >= 0 && msg < irsdk_BroadcastLast {
		sendNotifyMessage(msgID, msg2, wParam)
	}
}

// void irsdk_broadcastMsg(irsdk_BroadcastMsg msg, int var1, int var2)
// {
// 	static unsigned int msgId = irsdk_getBroadcastMsgID();

// 	if(msgId && msg >= 0 && msg < irsdk_BroadcastLast)
// 	{
// 		SendNotifyMessage(HWND_BROADCAST, msgId, MAKELONG(msg, var1), var2);
// 	}
// }

func irsdk_padCarNum(num int, zero int) int {
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
