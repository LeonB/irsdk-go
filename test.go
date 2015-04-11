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
		fmt.Println("checkered:", bool(flags&uint32(utils.Irsdk_checkered) != 0))
		fmt.Println("white:", bool(flags&uint32(utils.Irsdk_white) != 0))
		fmt.Println("green:", bool(flags&uint32(utils.Irsdk_green) != 0))
		fmt.Println("yellow:", bool(flags&uint32(utils.Irsdk_yellow) != 0))
		fmt.Println("red:", bool(flags&uint32(utils.Irsdk_red) != 0))
		fmt.Println("blue:", bool(flags&uint32(utils.Irsdk_blue) != 0))
		fmt.Println("debris:", bool(flags&uint32(utils.Irsdk_debris) != 0))
		fmt.Println("crossed:", bool(flags&uint32(utils.Irsdk_crossed) != 0))
		fmt.Println("yellowWaving:", bool(flags&uint32(utils.Irsdk_yellowWaving) != 0))
		fmt.Println("oneLapToGreen:", bool(flags&uint32(utils.Irsdk_oneLapToGreen) != 0))
		fmt.Println("greenHeld:", bool(flags&uint32(utils.Irsdk_greenHeld) != 0))
		fmt.Println("tenToGo:", bool(flags&uint32(utils.Irsdk_tenToGo) != 0))
		fmt.Println("fiveToGo:", bool(flags&uint32(utils.Irsdk_fiveToGo) != 0))
		fmt.Println("randomWaving:", bool(flags&uint32(utils.Irsdk_randomWaving) != 0))
		fmt.Println("caution:", bool(flags&uint32(utils.Irsdk_caution) != 0))
		fmt.Println("cautionWaving:", bool(flags&uint32(utils.Irsdk_cautionWaving) != 0))

		fmt.Println("black:", bool(flags&uint32(utils.Irsdk_black) != 0))
		fmt.Println("disqualify:", bool(flags&uint32(utils.Irsdk_disqualify) != 0))
		fmt.Println("servicible:", bool(flags&uint32(utils.Irsdk_servicible) != 0))
		fmt.Println("furled:", bool(flags&uint32(utils.Irsdk_furled) != 0))
		fmt.Println("repair:", bool(flags&uint32(utils.Irsdk_repair) != 0))

		fmt.Println("startHidden:", bool(flags&uint32(utils.Irsdk_startHidden) != 0))
		fmt.Println("startReady:", bool(flags&uint32(utils.Irsdk_startReady) != 0))
		fmt.Println("startSet:", bool(flags&uint32(utils.Irsdk_startSet) != 0))
		fmt.Println("startGo:", bool(flags&uint32(utils.Irsdk_startGo) != 0))
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
