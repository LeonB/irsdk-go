package main

import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
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
	irsdk_int = iota
	irsdk_bitField = iota
	irsdk_float = iota

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
var lastTickCount = INT_MAX

func main() {
	_, err := irsdk_startup()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connected: ", irsdk_isConnected())
	byteArray := irsdk_getSessionInfoStr()
	s := string(byteArray[:])
	fmt.Println("irsdk_getSessionInfoStr: ", s)
	// irsdk_getVarHeaderEntry(0)
	// irsdk_getVarHeaderPtr()

	varHeader := &irsdk_varHeader{}
	size := int(unsafe.Sizeof(*varHeader))

	fmt.Println(size)
	fmt.Println(pHeader.VarHeaderOffset)
	// fmt.Println(string(pSharedMem[pHeader.VarHeaderOffset:131484]))
	varHeaderOffset := int(pHeader.VarHeaderOffset)
	numVars := int(pHeader.NumVars)

	latest := 0
	for i := 0; i < int(pHeader.NumBuf); i++ {
		if pHeader.VarBuf[latest].TickCount < pHeader.VarBuf[i].TickCount {
			latest = i
		}
	}

	fmt.Println("latest: ", latest)

	for i := 0; i < int(pHeader.NumBuf); i++ {
		for count := 0; count < 2; count++ {
			curTickCount := int(pHeader.VarBuf[latest].TickCount)
			// memcpy(data, pSharedMem + pHeader->varBuf[latest].bufOffset, pHeader->bufLen)

			if curTickCount == int(pHeader.VarBuf[latest].TickCount) {
				fmt.Println("2")
				lastTickCount = curTickCount
				lastValidTime = time.Now()
			}
		}
	}

	for i := 0; i <= numVars; i++ {
		startByte := varHeaderOffset + (i * size)
		endByte := startByte + size

		b := bytes.NewBuffer(pSharedMem[startByte:endByte])
		// read []byte and convert it into irsd_header
		err = binary.Read(b, binary.LittleEndian, varHeader)

		// fmt.Printf("%+v\n", varHeader)
		// fmt.Println(string(varHeader.Name[:32]))
		// fmt.Println(varHeader.Type)
		// fmt.Println(string(varHeader.Desc[:]))
		// fmt.Println(string(varHeader.Unit[:]))
		// fmt.Println(varHeader.Count)

		// Type is also number of bytes?
		if varHeader.Type == irsdk_int {
			fmt.Println(string(varHeader.Name[:32]))
			// fprintf(file, "%s", (char *)(lineBuf+rec->offset) ); break;
		}
	}

}

func irsdk_startup() (bool, error) {
	var try winq.Try

	hMemMapFile := try.N("OpenFileMappingW", FILE_MAP_READ, false, syscall.StringToUTF16Ptr(IRSDK_MEMMAPFILENAME))
	if try.Err != nil {
		return false, try.Err
	}

	sharedMemPtr := try.N("MapViewOfFile", hMemMapFile, syscall.FILE_MAP_READ, 0, 0, 0)
	pSharedMem = (*[1 << 30]byte)(unsafe.Pointer(sharedMemPtr))[:]

	// create a io.Reader
	b := bytes.NewBuffer(pSharedMem[:unsafe.Sizeof(*pHeader)])
	// read []byte and convert it into irsd_header
	pHeader = &irsdk_header{}
	err := binary.Read(b, binary.LittleEndian, pHeader)
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
