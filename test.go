package main

import "C"
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/akavel/winq"
)

const (
	IRSDK_MEMMAPFILENAME     = "Local\\IRSDKMemMapFileName"
	IRSDK_DATAVALIDEVENTNAME = "Local\\IRSDKDataValidEvent"
	INT_MAX                  = 2147483647
	SYNCHRONIZE              = 1048576

	IRSDK_MAX_BUFS   = 4
	IRSDK_MAX_STRING = 32
	// descriptions can be longer than max_string!
	IRSDK_MAX_DESC = 64

	irsdk_stConnected = 1
	TIMEOUT           = time.Duration(30) // timeout after 30 seconds with no communication
)

const (
	// 1 byte
	irsdk_char = iota
	irsdk_bool = iota

	// 4 bytes
	irsdk_int      = iota
	irsdk_bitField = iota
	irsdk_float    = iota

	// 8 bytes
	irsdk_double = iota
)

type irsdk_varBuf struct {
	TickCount C.int    // used to detect changes in data
	BufOffset C.int    // offset from header
	Pad       [2]C.int // (16 byte align)
}

type irsdk_header struct {
	Ver      C.int // api version 1 for now
	Status   C.int // bitfield using irsdk_StatusField
	TickRate C.int // ticks per second (60 or 360 etc)

	// session information, updated periodicaly
	SessionInfoUpdate C.int // Incremented when session info changes
	SessionInfoLen    C.int // Length in bytes of session info string
	SessionInfoOffset C.int // Session info, encoded in YAML format

	// State data, output at tickRate
	NumVars         C.int // length of array pointed to by varHeaderOffset
	VarHeaderOffset C.int // offset to irsdk_varHeader[numVars] array, Describes the variables recieved in varBuf

	NumBuf C.int    // <= IRSDK_MAX_BUFS (3 for now)
	BufLen C.int    // length in bytes for one line
	Pad1   [2]C.int // (16 byte align)
	VarBuf [IRSDK_MAX_BUFS]irsdk_varBuf
}

type irsdk_varHeader struct {
	Type   C.int // irsdk_VarType
	Offset C.int // offset fron start of buffer row
	Count  C.int // number of entrys (array)
	// so length in bytes would be irsdk_VarTypeBytes[type] * count

	Pad [1]C.int // (16 byte align)

	Name [IRSDK_MAX_STRING]byte
	Desc [IRSDK_MAX_DESC]byte
	Unit [IRSDK_MAX_STRING]byte // something like "kg/m^2"
}

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

func main() {
	var data []byte

	result := false
	for result == false {
		result = irsdk_getNewData(data)

		if result == true {
			fmt.Println("data:", data)
			fmt.Println("len(data): ", len(data))
			numVars := int(pHeader.NumVars)

			for i := 0; i <= numVars; i++ {
				varHeader := irsdk_getVarHeaderEntry(i)

				if varHeader != nil {
					// fmt.Println("varHeader.Offset: ", varHeader.Offset)

					if varHeader.Type == irsdk_int {
						var myvar C.int
						count := int(varHeader.Count)
						startByte := int(varHeader.Offset)
						varLen := int(unsafe.Sizeof(myvar))
						endByte := startByte + varLen
						fmt.Println("varHeader.Name:", CToGoString(varHeader.Name[:]))
						fmt.Println("count:", count)
						fmt.Println("type:", "int")

						buf := bytes.NewBuffer(data[startByte:endByte])
						binary.Read(buf, binary.LittleEndian, &myvar)
						fmt.Println("myvar: ", myvar)
					} else if varHeader.Type == irsdk_float {
						var myvar C.float
						count := int(varHeader.Count)
						startByte := int(varHeader.Offset)
						varLen := int(unsafe.Sizeof(myvar))
						endByte := startByte + varLen
						fmt.Println("varHeader.Name:", CToGoString(varHeader.Name[:]))
						fmt.Println("count: ", count)
						fmt.Println("type:", "float")

						buf := bytes.NewBuffer(data[startByte:endByte])
						binary.Read(buf, binary.LittleEndian, &myvar)
						fmt.Println("myvar: ", myvar)
					}
				}
			}
		}

		// try.N("Sleep", 10)
	}

	irsdk_shutdown()
	irsdk_startup()

	// byteArray := irsdk_getSessionInfoStr()
	// s := string(byteArray[:])
	// fmt.Println("irsdk_getSessionInfoStr: ", s)

	return

}

func irsdk_startup() (bool, error) {
	var try winq.Try

	if hMemMapFile == 0 {
		hMemMapFile = try.N("OpenFileMappingW", syscall.FILE_MAP_READ, false, syscall.StringToUTF16Ptr(IRSDK_MEMMAPFILENAME))
		if try.Err != nil {
			return false, try.Err
		}
		lastTickCount = INT_MAX
	}

	if hMemMapFile != 0 {
		if len(pSharedMem) == 0 {
			sharedMemPtr := try.N("MapViewOfFile", hMemMapFile, syscall.FILE_MAP_READ, 0, 0, 0)
			pHeader = (*irsdk_header)(unsafe.Pointer(sharedMemPtr))
			pSharedMem = (*[1 << 30]byte)(unsafe.Pointer(sharedMemPtr))[:]
			lastTickCount = INT_MAX
		}

		if len(pSharedMem) != 0 {
			if hDataValidEvent == 0 {
				hDataValidEvent = try.N("OpenEvent", SYNCHRONIZE, false, syscall.StringToUTF16Ptr(IRSDK_DATAVALIDEVENTNAME))
				lastTickCount = INT_MAX
			}

			if hDataValidEvent != 0 {
				isInitialized = true
				return isInitialized, nil
			}
			//else printf("Error opening event: %d\n", GetLastError());
		}
		//else printf("Error mapping file: %d\n", GetLastError());
	}
	//else printf("Error opening file: %d\n", GetLastError()); `

	isInitialized = false
	return isInitialized, errors.New("Failed to initialize")
}

func irsdk_shutdown() {
	var try winq.Try

	if hDataValidEvent != 0 {
		try.N("CloseHandle", hDataValidEvent)

		if len(pSharedMem) != 0 {
			sharedMemPtr := uintptr(unsafe.Pointer(&pSharedMem))
			try.N("UnmapViewOfFile", sharedMemPtr)

			if hMemMapFile != 0 {
				try.N("CloseHandle", hMemMapFile)

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
func irsdk_getNewData(data []byte) (bool) {
	returnData := true

	fmt.Println(unsafe.Pointer(&data))
	data = []byte("hellø")
	fmt.Println(unsafe.Pointer(&data))
	fmt.Println("data:", data)
	fmt.Println("len(data): ", len(data))
	return true

	if !isInitialized {
		success, _ := irsdk_startup()
		if !success {
			return false
		}
	}

	// if sim is not active, then no new data
	if (int(pHeader.Status) & irsdk_stConnected) == 0 {
		lastTickCount = INT_MAX
		return false
	}

	latest := 0
	for i := 0; i < int(pHeader.NumBuf); i++ {
		if pHeader.VarBuf[latest].TickCount < pHeader.VarBuf[i].TickCount {
			latest = i
		}
	}

	// if newer than last recieved, than report new data
	if lastTickCount < int(pHeader.VarBuf[latest].TickCount) {

		// if asked to retrieve the data
		if returnData == true {

			for count := 0; count < 2; count++ {
				curTickCount := int(pHeader.VarBuf[latest].TickCount)
				bufLen := int(pHeader.BufLen)
				startByte := int(pHeader.VarBuf[latest].BufOffset)
				endByte := startByte + bufLen

				fmt.Println("startByte: ", startByte)
				fmt.Println("endByte: ", endByte)

				// memcpy(data, pSharedMem + pHeader->varBuf[latest].bufOffset, pHeader->bufLen)

				data = make([]byte, bufLen)
				copy(data, pSharedMem[startByte:endByte])
				data = pSharedMem[startByte:endByte]

				fmt.Println("bufLen: ", bufLen)
				fmt.Println("len(data): ", len(data))

				if curTickCount == int(pHeader.VarBuf[latest].TickCount) {
					lastTickCount = curTickCount
					lastValidTime = time.Now()
					return true
				}
			}
			// if here, the data changed out from under us.
			return false
		} else {
			lastTickCount = int(pHeader.VarBuf[latest].TickCount)
			lastValidTime = time.Now()
			return true
		}
	} else if lastTickCount > int(pHeader.VarBuf[latest].TickCount) {
		// if older than last recieved, than reset, we probably disconnected
		lastTickCount = int(pHeader.VarBuf[latest].TickCount)
		return false
	}
	// else the same, and nothing changed this tick

	return false
}

func irsdk_waitForDataReady(timeOut int, data []byte) bool {
	var try winq.Try

	if !isInitialized {
		success, _ := irsdk_startup()

		if !success {
			// sleep if error
			// @TODO: fix this
			if timeOut > 0 {
				try.N("Sleep", timeout)
			}

			return false
		}
	}

	// just to be sure, check before we sleep
	if irsdk_getNewData(data) {
		return true
	}

	// sleep till signaled
	try.N("WaitForSingleObject", hDataValidEvent, timeOut)

	// we woke up, so check for data
	if irsdk_getNewData(data) {
		return true
	} else {
		return false
	}

	// sleep if error
	if timeOut > 0 {
		try.N("Sleep", timeOut)
	}

	return false
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

// unsigned int irsdk_getBroadcastMsgID()
// {
// 	static unsigned int msgId = RegisterWindowMessageA(IRSDK_BROADCASTMSGNAME);

// 	return msgId;
// }

// void irsdk_broadcastMsg(irsdk_BroadcastMsg msg, int var1, int var2, int var3)
// {
// 	irsdk_broadcastMsg(msg, var1, MAKELONG(var2, var3));
// }

// void irsdk_broadcastMsg(irsdk_BroadcastMsg msg, int var1, int var2)
// {
// 	static unsigned int msgId = irsdk_getBroadcastMsgID();

// 	if(msgId && msg >= 0 && msg < irsdk_BroadcastLast)
// 	{
// 		SendNotifyMessage(HWND_BROADCAST, msgId, MAKELONG(msg, var1), var2);
// 	}
// }

// int irsdk_padCarNum(int num, int zero)
// {
// 	int retVal = num;
// 	int numPlace = 1;
// 	if(num > 99)
// 		numPlace = 3;
// 	else if(num > 9)
// 		numPlace = 2;
// 	if(zero)
// 	{
// 		numPlace += zero;
// 		retVal = num + 1000*numPlace;
// 	}

// 	return retVal;
// }
