// +build windows

package main

import "C"
import (
	"fmt"
)

func main() {

	err := irsdk_broadcastMsg(
		irsdk_BroadcastChatComand,
		uint16(irsdk_ChatCommand_BeginChat),
		0,
		0,
	)
	if err != nil {
		fmt.Println(err)
	}

	chatMacro := 1
	fmt.Printf("Sending chat macro %d\n", chatMacro)
	err = irsdk_broadcastMsg(
		irsdk_BroadcastChatComand,
		uint16(irsdk_ChatCommand_Macro),
		uint16(chatMacro),
		0,
	);
	if err != nil {
		fmt.Println(err)
	}
	return

	// var data []byte
	// var err error

	// oldTime := time.Now().Unix()
	// changes := 0
	// for {
	// 	newTime := time.Now().Unix()
	// 	// fmt.Println(newTime)

	// 	if oldTime != newTime {
	// 		oldTime = newTime
	// 		changes = 0
	// 		fmt.Println("number of changes:", changes)
	// 	}

	// 	data, err = irsdk_waitForDataReady(1000)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}

	// 	if data != nil {
	// 		fmt.Println("Data changed")
	// 		changes++
	// 		// testData(data)
	// 	}

	// 	// irsdk_shutdown()
	// 	// break
	// }

	// return
// }

// func testData(data []byte) {
	// fmt.Println("data:", data)
	// fmt.Println("len(data): ", len(data))
	// numVars := int(pHeader.NumVars)

	// for i := 0; i <= numVars; i++ {
	// 	varHeader := irsdk_getVarHeaderEntry(i)

	// 	if varHeader != nil {
	// 		// fmt.Println("varHeader.Offset: ", varHeader.Offset)

	// 		if varHeader.Type == irsdk_int {
	// 			var myvar C.int
	// 			count := int(varHeader.Count)
	// 			startByte := int(varHeader.Offset)
	// 			varLen := int(unsafe.Sizeof(myvar))
	// 			endByte := startByte + varLen
	// 			fmt.Println("varHeader.Name:", CToGoString(varHeader.Name[:]))
	// 			fmt.Println("count:", count)
	// 			fmt.Println("type:", "int")

	// 			buf := bytes.NewBuffer(data[startByte:endByte])
	// 			binary.Read(buf, binary.LittleEndian, &myvar)
	// 			fmt.Println("myvar: ", myvar)
	// 		} else if varHeader.Type == irsdk_float {
	// 			var myvar C.float
	// 			count := int(varHeader.Count)
	// 			startByte := int(varHeader.Offset)
	// 			varLen := int(unsafe.Sizeof(myvar))
	// 			endByte := startByte + varLen
	// 			fmt.Println("varHeader.Name:", CToGoString(varHeader.Name[:]))
	// 			fmt.Println("count: ", count)
	// 			fmt.Println("type:", "float")

	// 			buf := bytes.NewBuffer(data[startByte:endByte])
	// 			binary.Read(buf, binary.LittleEndian, &myvar)
	// 			fmt.Println("myvar: ", myvar)
	// 		}
	// 	}
	// }
}
