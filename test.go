// +build windows

package main

import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	utils "github.com/leonb/irsdk-go/irsdk_utils"
)

type irCharVar struct {
	name  string
	value byte
}

type irBoolVar struct {
	name  string
	value bool
}

type irIntVar struct {
	name  string
	value int
}

type irBitfieldVar struct {
	name string
	flags map[string]bool
}

type irFloatVar struct {
	name  string
	value float32
}

func main() {
	// testBroadcastMsg()
	testTelemetryData()
}

func testTelemetryData() {
	var data []byte
	var err error

	// oldTime := time.Now().Unix()
	changes := 0
	for {
		// newTime := time.Now().Unix()
		// fmt.Println(newTime)

		// if oldTime != newTime {
		// 	oldTime = newTime
		// 	changes = 0
		// 	fmt.Println("number of changes:", changes)
		// }

		data, err = utils.Irsdk_waitForDataReady(72)
		if err != nil {
			fmt.Println(err)
		}

		if data != nil {
			fmt.Println("Data changed")
			changes++
			testData(data)
			break
		}

		// irsdk_shutdown()
		// break
	}

	return
}

func testData(data []byte) {
	// fmt.Println("data:", data)
	// fmt.Println("len(data): ", len(data))
	numVars := utils.Irsdk_getNumVars()

	for i := 0; i <= numVars; i++ {
		varHeader := utils.Irsdk_getVarHeaderEntry(i)

		if varHeader != nil {
			// fmt.Println("varHeader.Offset: ", varHeader.Offset)

			// make this a switch

			switch varHeader.Type {
			case utils.Irsdk_char:
				continue
				extractCharFromVarHeader(varHeader, data)
			case utils.Irsdk_bool:
				continue
				extractBoolFromVarHeader(varHeader, data)
			case utils.Irsdk_int:
				continue
				extractIntFromVarHeader(varHeader, data)
			case utils.Irsdk_bitField:
				extractBitfieldFromVarHeader(varHeader, data)
			case utils.Irsdk_float:
				continue
				extractFloatFromVarHeader(varHeader, data)
			}
		}
	}
}

func extractCharFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irCharVar {
	var hvar C.char // 1 byte
	retVar := irCharVar{}

	count := int(header.Count)
	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	fmt.Println("header.Name:", utils.CToGoString(header.Name[:]))
	fmt.Println("count:", count)
	fmt.Println("type:", "char")

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)
	fmt.Println("hvar: ", hvar)

	return retVar
}

func extractBoolFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irBoolVar {
	var hvar bool // 1 byte
	retVar := irBoolVar{}

	count := int(header.Count)
	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	fmt.Println("header.Name:", utils.CToGoString(header.Name[:]))
	fmt.Println("count:", count)
	fmt.Println("type:", "bool")

	if data[startByte:endByte][0] == 0 {
		hvar = false
	} else {
		hvar = true
	}

	fmt.Println("hvar: ", hvar)

	return retVar
}

func extractIntFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irIntVar {
	var hvar C.int // 4 bytes
	retVar := irIntVar{}

	count := int(header.Count)
	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	fmt.Println("header.Name:", utils.CToGoString(header.Name[:]))
	fmt.Println("count:", count)
	fmt.Println("type:", "int")

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)
	fmt.Println("hvar: ", hvar)

	return retVar
}

func extractBitfieldFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irBitfieldVar {
	// - SessionFlags
	// - CamCameraState
	// - EngineWarnings
	var hvar uint32 // 4 bytes
	retVar := irBitfieldVar{}

	count := int(header.Count)
	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	fmt.Println("header.Name:", varName)
	fmt.Println("count:", count)
	fmt.Println("type:", "bitField")

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)
	fmt.Println("hvar: ", hvar)
	fmt.Println("data[startByte:endByte]:", data[startByte:endByte])

	if varName == "SessionFlags" {
		flags := uint32(hvar)
		fmt.Println("checkered:", flags&uint32(utils.Irsdk_checkered))
		fmt.Println("white:", flags&uint32(utils.Irsdk_white))
		fmt.Println("green:", flags&uint32(utils.Irsdk_green))
		fmt.Println("yellow:", flags&uint32(utils.Irsdk_yellow))
		fmt.Println("red:", flags&uint32(utils.Irsdk_red))
		fmt.Println("blue:", flags&uint32(utils.Irsdk_blue))
		fmt.Println("debris:", flags&uint32(utils.Irsdk_debris))
		fmt.Println("crossed:", flags&uint32(utils.Irsdk_crossed))
		fmt.Println("yellowWaving:", flags&uint32(utils.Irsdk_yellowWaving))
		fmt.Println("oneLapToGreen:", bool(flags&uint32(utils.Irsdk_oneLapToGreen) != 0))
		fmt.Println("greenHeld", flags&uint32(utils.Irsdk_greenHeld))
		fmt.Println("tenToGo", flags&uint32(utils.Irsdk_tenToGo))
		fmt.Println("fiveToGo", flags&uint32(utils.Irsdk_fiveToGo))
		fmt.Println("randomWaving", flags&uint32(utils.Irsdk_randomWaving))
		fmt.Println("caution", flags&uint32(utils.Irsdk_caution))
		fmt.Println("cautionWaving", flags&uint32(utils.Irsdk_cautionWaving))

		fmt.Println("black", flags&uint32(utils.Irsdk_black))
		fmt.Println("disqualify", flags&uint32(utils.Irsdk_disqualify))
		fmt.Println("servicible", flags&uint32(utils.Irsdk_servicible))
		fmt.Println("furled", flags&uint32(utils.Irsdk_furled))
		fmt.Println("repair", flags&uint32(utils.Irsdk_repair))

		fmt.Println("startHidden", flags&uint32(utils.Irsdk_startHidden))
		fmt.Println("startReady", flags&uint32(utils.Irsdk_startReady))
		fmt.Println("startSet", flags&uint32(utils.Irsdk_startSet))
		fmt.Println("startGo", flags&uint32(utils.Irsdk_startGo))
	}

	return retVar
}

func extractFloatFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irFloatVar {
	var hvar C.float
	retVar := irFloatVar{}

	count := int(header.Count)
	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	fmt.Println("header.Name:", utils.CToGoString(header.Name[:]))
	fmt.Println("count: ", count)
	fmt.Println("type:", "float")

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)
	fmt.Println("hvar: ", hvar)

	return retVar
}

func testBroadcastMsg() {
	err := utils.Irsdk_broadcastMsg(
		utils.Irsdk_BroadcastChatComand,
		uint16(utils.Irsdk_ChatCommand_BeginChat),
		0,
		0,
	)
	if err != nil {
		fmt.Println(err)
	}

	chatMacro := 1
	fmt.Printf("Sending chat macro %d\n", chatMacro)
	err = utils.Irsdk_broadcastMsg(
		utils.Irsdk_BroadcastChatComand,
		uint16(utils.Irsdk_ChatCommand_Macro),
		uint16(chatMacro),
		0,
	)
	if err != nil {
		fmt.Println(err)
	}
	return
}
