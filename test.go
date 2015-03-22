package main

import "C"
import (
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
	SYNCHRONIZE              = 1048576

	IRSDK_MAX_BUFS   = 4
	IRSDK_MAX_STRING = 32
	// descriptions can be longer than max_string!
	IRSDK_MAX_DESC = 64

	irsdk_stConnected = 1
	TIMEOUT           = time.Duration(30) // timeout after 30 seconds with no communication
)

type irsdk_varBuf struct {
	tickCount C.int    // used to detect changes in data
	bufOffset C.int    // offset from header
	pad       [2]C.int // (16 byte align)
}

type irsdk_header struct {
	ver      C.int // api version 1 for now
	status   C.int // bitfield using irsdk_StatusField
	tickRate C.int // ticks per second (60 or 360 etc)

	// session information, updated periodicaly
	sessionInfoUpdate C.int // Incremented when session info changes
	sessionInfoLen    C.int // Length in bytes of session info string
	sessionInfoOffset C.int // Session info, encoded in YAML format

	// State data, output at tickRate

	numVars         C.int // length of array pointed to by varHeaderOffset
	varHeaderOffset C.int // offset to irsdk_varHeader[numVars] array, Describes the variables recieved in varBuf

	numBuf C.int // <= IRSDK_MAX_BUFS (3 for now)
	bufLen C.int // length in bytes for one line
	pad    [1]C.int
	varBuf [IRSDK_MAX_BUFS]irsdk_varBuf
}

type irsdk_varHeader struct {
	Type   C.int // irsdk_VarType
	offset C.int // offset fron start of buffer row
	count  C.int // number of entrys (array)
	// so length in bytes would be irsdk_VarTypeBytes[type] * count
	pad [1]C.int // (16 byte align)

	name [IRSDK_MAX_STRING]byte
	desc [IRSDK_MAX_DESC]byte
	unit [IRSDK_MAX_STRING]byte // something like "kg/m^2"
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
var pSharedMem unsafe.Pointer

func main() {
	irsdk_startup()
	fmt.Println("connected: ", irsdk_isConnected())
	byteArray := irsdk_getSessionInfoStr()
	s := string(byteArray[:])
	fmt.Println("irsdk_getSessionInfoStr: ", s)
	// irsdk_getVarHeaderEntry(0)
	// irsdk_getVarHeaderPtr()
}

func irsdk_startup() (bool, error) {
	var try winq.Try

	hMemMapFile := try.N("OpenFileMappingW", FILE_MAP_READ, false, syscall.StringToUTF16Ptr(IRSDK_MEMMAPFILENAME))
	if try.Err != nil {
		return false, try.Err
	}

	pSharedMem = unsafe.Pointer(uintptr(try.N("MapViewOfFile", hMemMapFile, FILE_MAP_READ, 0, 0, 0)))
	pHeader = (*irsdk_header)(pSharedMem)

	fmt.Println("pHeader.ver: ", pHeader.ver)
	fmt.Println("pHeader.status: ", pHeader.status)
	fmt.Println("pHeader.sessionInfoOffset: ", pHeader.sessionInfoOffset)
	fmt.Println("pHeader.sessionInfoUpdate: ", pHeader.sessionInfoUpdate)

	hDataValidEvent := try.N("OpenEvent", SYNCHRONIZE, false, syscall.StringToUTF16Ptr(IRSDK_DATAVALIDEVENTNAME))

	if hDataValidEvent > 0 {
		isInitialized = true
		return isInitialized, nil
	}

	println("got:", pSharedMem)
	println("got:", hDataValidEvent)
	return false, nil
}

func irsdk_getSessionInfoStr() []byte {
	if isInitialized {

		// fmt.Println(C.GoString((*C.char)(pSharedMem)))
		byteSlice := C.GoBytes(pSharedMem, pHeader.sessionInfoLen)
		return byteSlice[pHeader.sessionInfoOffset:]
		return C.GoBytes(pSharedMem, pHeader.sessionInfoOffset)
	}
	return nil
}

func irsdk_isConnected() bool {
	if isInitialized {
		elapsed := time.Now().Sub(lastValidTime)
		fmt.Println("elapsed: ", elapsed)
		fmt.Println("timeout: ", timeout)
		fmt.Println("pHeader.status&irsdk_stConnected: ", pHeader.status&irsdk_stConnected)
		if (pHeader.status&irsdk_stConnected) > 0 && (elapsed < timeout) {
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
