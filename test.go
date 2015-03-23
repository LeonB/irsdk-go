package main

import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/akavel/winq"
)

const (
	IRSDK_MEMMAPFILENAME     = "Local\\IRSDKMemMapFileName"
	IRSDK_DATAVALIDEVENTNAME = "Local\\IRSDKDataValidEvent"
	SECTION_MAP_READ         = 4
	FILE_MAP_READ            = SECTION_MAP_READ
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

	NumBuf C.int // <= IRSDK_MAX_BUFS (3 for now)
	BufLen C.int // length in bytes for one line
	Pad    [1]C.int
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

func irsdk_getVarHeaderPtr() *irsdk_varHeader {
	// if isInitialized {
	// 	varHeaderoffset := unsafe.Sizeof(pHeader.varHeaderOffset)
	// 	varHeader := (*irsdk_varHeader)(unsafe.Pointer(pSharedMem) + varHeaderoffset)

	// 	fmt.Printf("%+v\n", varHeader)
	// 	fmt.Println(string(varHeader.name[:10]))
	// 	return nil
	// }
	return nil
}

// func irsdk_getVarHeaderEntry(index int) *irsdk_varHeader {
// 	if isInitialized {
// 		if index >= 0 && index < (int)(pHeader.numVars) {

// 			varHeader := (*irsdk_varHeader)((unsafe.Pointer((uintptr(pSharedMem) + unsafe.Sizeof(pHeader.varHeaderOffset)))))
// 			fmt.Printf("%+v\n", varHeader.name)
// 			fmt.Println(string(varHeader.name[:10]))
// 			fmt.Println(string(varHeader.desc[:10]))
// 			fmt.Println(string(varHeader.unit[:10]))
// 			return nil
// 			// return &((irsdk_varHeader*)(pSharedMem + pHeader->varHeaderOffset))[index];
// 		}
// 	}
// 	return nil
// }

var pHeader *irsdk_header
var isInitialized bool
var lastValidTime time.Time
var timeout time.Duration
var pSharedMem []byte
var sharedMemPtr uintptr
var lastTickCount = INT_MAX

func main() {
	// _, err := irsdk_startup()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	for i := 0; i < 90000; i++ {
		data, success := irsdk_getNewData()
		fmt.Println("success: ", success)
		fmt.Println("len(data): ", len(data))
	}

	return

	// fmt.Println("connected: ", irsdk_isConnected())
	// byteArray := irsdk_getSessionInfoStr()
	// s := string(byteArray[:])
	// fmt.Println("irsdk_getSessionInfoStr: ", s)

}

func irsdk_startup() (bool, error) {
	var try winq.Try

	hMemMapFile := try.N("OpenFileMappingW", FILE_MAP_READ, false, syscall.StringToUTF16Ptr(IRSDK_MEMMAPFILENAME))
	if try.Err != nil {
		return false, try.Err
	}

	sharedMemPtr = try.N("MapViewOfFile", hMemMapFile, syscall.FILE_MAP_READ, 0, 0, 0)

	err := updateSharedMem()
	if err != nil {
		return false, err
	}

	hDataValidEvent := try.N("OpenEvent", SYNCHRONIZE, false, syscall.StringToUTF16Ptr(IRSDK_DATAVALIDEVENTNAME))

	if hDataValidEvent > 0 {
		isInitialized = true
		return isInitialized, nil
	}

	return false, nil
}

func irsdk_getSessionInfoStr() []byte {
	if isInitialized {
		return pSharedMem[pHeader.SessionInfoOffset:pHeader.SessionInfoLen]
	}
	return nil
}

func irsdk_isConnected() bool {
	if isInitialized {
		elapsed := time.Now().Sub(lastValidTime)
		fmt.Println("elapsed: ", elapsed)
		fmt.Println("timeout: ", timeout)
		fmt.Println("pHeader.status&irsdk_stConnected: ", pHeader.Status&irsdk_stConnected)
		if (pHeader.Status&irsdk_stConnected) > 0 && (elapsed < timeout) {
			return true
		}
	}

	return false
}

// func irsdk_varHeader *irsdk_getVarHeaderPtr()
// {
// 	if(isInitialized)
// 	{
// 		return ((irsdk_varHeader*)(pSharedMem + pHeader->varHeaderOffset));
// 	}
// 	return NULL;
// }

func irsdk_getNewData() ([]byte, bool) {
	var data []byte
	returnData := true

	if !isInitialized {
		success, _ := irsdk_startup()
		if !success {
			return nil, false
		}
	}

	// if sim is not active, then no new data
	if (int(pHeader.Status) & irsdk_stConnected) == 0 {
		lastTickCount = INT_MAX
		return nil, false
	}

	latest := 0
	for i := 0; i < int(pHeader.NumBuf); i++ {
		fmt.Println("i: ", i)
		fmt.Println("pHeader.VarBuf[latest].TickCount: ", pHeader.VarBuf[latest].TickCount)
		fmt.Println("pHeader.VarBuf[i].TickCount: ", pHeader.VarBuf[i].TickCount)
		if pHeader.VarBuf[latest].TickCount < pHeader.VarBuf[i].TickCount {
			latest = i
		}
	}

	fmt.Println("latest: ", latest)

	fmt.Println("lastTickCount: ", lastTickCount)
	fmt.Println("pHeader.VarBuf[latest].TickCount: ", int(pHeader.VarBuf[latest].TickCount))

	// if newer than last recieved, than report new data
	if lastTickCount < int(pHeader.VarBuf[latest].TickCount) {
		// if asked to retrieve the data
		if returnData == true {

			for count := 0; count < 2; count++ {
				curTickCount := int(pHeader.VarBuf[latest].TickCount)
				// memcpy(data, pSharedMem + pHeader->varBuf[latest].bufOffset, pHeader->bufLen)
				startByte := int(pHeader.VarBuf[latest].BufOffset)
				startByte = 1358715
				endByte := startByte + int(pHeader.BufLen)
				fmt.Println("startByte: ", startByte)
				fmt.Println("endByte: ", endByte)
				fmt.Println("len(pSharedMem): ", len(pSharedMem))
				data = pSharedMem[startByte:endByte]
				fmt.Println("len(data): ", len(data))
				// fmt.Println("data: ", string(data[:]))

				if curTickCount == int(pHeader.VarBuf[latest].TickCount) {
					lastTickCount = curTickCount
					lastValidTime = time.Now()
					return data, true
				}
			}
			// if here, the data changed out from under us.
			return nil, false
		} else {
			lastTickCount = int(pHeader.VarBuf[latest].TickCount)
			lastValidTime = time.Now()
			return data, true
		}
	} else if lastTickCount > int(pHeader.VarBuf[latest].TickCount) {
		// if older than last recieved, than reset, we probably disconnected
		lastTickCount = int(pHeader.VarBuf[latest].TickCount)
		return nil, false
	}
	// else the same, and nothing changed this tick

	return nil, false
}

func updateSharedMem() error {
	pHeader = (*irsdk_header)(unsafe.Pointer(sharedMemPtr))

	pSharedMem = (*[1 << 30]byte)(unsafe.Pointer(sharedMemPtr))[:]
	return nil

	// This is also an option:

	// create a io.Reader
	endByte := unsafe.Sizeof(*pHeader)
	b := bytes.NewBuffer(pSharedMem[:endByte])
	// read []byte and convert it into irsd_header
	pHeader = &irsdk_header{}
	err := binary.Read(b, binary.LittleEndian, pHeader)
	return err
}
