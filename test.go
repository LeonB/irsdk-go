package main

import (
	"C"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"unsafe"
)

func main() {
	var data []byte
	var err error

	result := false
	for result == false {
		data, err = irsdk_waitForDataReady(1000)
		if err != nil {
			log.Println(err)
		}

		if data != nil {
			fmt.Println("Data changed")
			// testData(data)
		}

		irsdk_shutdown()
		break
	}

	return
}

func testData(data []byte) {
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
