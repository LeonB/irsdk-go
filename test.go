// +build windows

package main

import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"unsafe"

	utils "github.com/leonb/irsdk-go/irsdk_utils"
)

type irCharVar struct {
	name  string
	desc  string
	value byte
	unit  string
}

type irBoolVar struct {
	name  string
	desc  string
	value bool
	unit  string
}

type irIntVar struct {
	name  string
	desc  string
	value int
	unit  string
}

type irBitfieldVar struct {
	name   string
	desc   string
	fields map[string]bool
	unit   string
}

type irFloatVar struct {
	name  string
	desc  string
	value float32
	unit  string
}

type irDoubleVar struct {
	name  string
	desc  string
	value float64
	unit  string
}

var irsdkFlags = map[utils.Irsdk_Flags]string{
	// global flags
	utils.Irsdk_checkered:     "checkered",
	utils.Irsdk_white:         "white",
	utils.Irsdk_green:         "green",
	utils.Irsdk_yellow:        "yellow",
	utils.Irsdk_red:           "red",
	utils.Irsdk_blue:          "blue",
	utils.Irsdk_debris:        "debris",
	utils.Irsdk_crossed:       "crossed",
	utils.Irsdk_yellowWaving:  "yellowWaving",
	utils.Irsdk_oneLapToGreen: "oneLapToGreen",
	utils.Irsdk_greenHeld:     "greenHeld",
	utils.Irsdk_tenToGo:       "tenToGo",
	utils.Irsdk_fiveToGo:      "fiveToGo",
	utils.Irsdk_randomWaving:  "randomWaving",
	utils.Irsdk_caution:       "caution",
	utils.Irsdk_cautionWaving: "cautionWaving",

	// drivers black flags
	utils.Irsdk_black:      "black",
	utils.Irsdk_disqualify: "disqualify",
	utils.Irsdk_servicible: "servicible", // car is allowed service (not a flag)
	utils.Irsdk_furled:     "furled",
	utils.Irsdk_repair:     "repair",

	// start lights
	utils.Irsdk_startHidden: "startHidden",
	utils.Irsdk_startReady:  "startReady",
	utils.Irsdk_startSet:    "startSet",
	utils.Irsdk_startGo:     "startGo",
}

var irsdkEngineWarnings = map[utils.Irsdk_EngineWarnings]string{
	utils.Irsdk_waterTempWarning:    "waterTempWarning",
	utils.Irsdk_fuelPressureWarning: "fuelPressureWarning",
	utils.Irsdk_oilPressureWarning:  "oilPressureWarning",
	utils.Irsdk_engineStalled:       "engineStalled",
	utils.Irsdk_pitSpeedLimiter:     "pitSpeedLimiter",
	utils.Irsdk_revLimiterActive:    "revLimiterActive",
}

var irsdkCameraStates = map[utils.Irsdk_CameraState]string{
	utils.Irsdk_IsSessionScreen:       "isSessionScreen",
	utils.Irsdk_IsScenicActive:        "isScencActive",
	utils.Irsdk_CamToolActive:         "camToolActive",
	utils.Irsdk_UIHidden:              "uiHidden",
	utils.Irsdk_UseAutoShotSelection:  "useAutoShotSelection",
	utils.Irsdk_UseTemporaryEdits:     "useTemporaryEdits",
	utils.Irsdk_UseKeyAcceleration:    "useKeyAcceleration",
	utils.Irsdk_UseKey10xAcceleration: "useKey10xAcceleration",
	utils.Irsdk_UseMouseAimMode:       "useMouseAimMode",
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
		}

		utils.Irsdk_shutdown()
		break
	}

	return
}

func testData(data []byte) {
	numVars := utils.Irsdk_getNumVars()

	for i := 0; i <= numVars; i++ {
		varHeader := utils.Irsdk_getVarHeaderEntry(i)

		if varHeader != nil {
			switch varHeader.Type {
			case utils.Irsdk_char:
				irVar := extractCharFromVarHeader(varHeader, data)
				fmt.Printf("%v: %v\n", irVar.name, irVar.value)
				fmt.Println(irVar.desc)
			case utils.Irsdk_bool:
				irVar := extractBoolFromVarHeader(varHeader, data)
				fmt.Printf("%v: %v\n", irVar.name, irVar.value)
				fmt.Println(irVar.desc)
			case utils.Irsdk_int:
				irVar := extractIntFromVarHeader(varHeader, data)
				fmt.Printf("%v: %v\n", irVar.name, irVar.value)
				fmt.Println(irVar.desc)
			case utils.Irsdk_bitField:
				irVar := extractBitfieldFromVarHeader(varHeader, data)
				fmt.Printf("%v: %v\n", irVar.name, irVar.fields)
				fmt.Println(irVar.desc)
			case utils.Irsdk_float:
				irVar := extractFloatFromVarHeader(varHeader, data)
				fmt.Printf("%v: %v\n", irVar.name, irVar.value)
				fmt.Println(irVar.desc)
			case utils.Irsdk_double:
				irVar := extractDoubleFromVarHeader(varHeader, data)
				fmt.Printf("%v: %v\n", irVar.name, irVar.value)
				fmt.Println(irVar.desc)
			default:
				log.Println("Unknown irsdk varType:", varHeader.Type)
			}
		}
	}
}

func extractCharFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irCharVar {
	var hvar C.char // 1 byte

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	return irCharVar{
		name:  varName,
		desc:  varDesc,
		value: byte(hvar),
		unit:  varUnit,
	}
}

func extractBoolFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irBoolVar {
	var hvar bool // 1 byte

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	if data[startByte:endByte][0] == 0 {
		hvar = false
	} else {
		hvar = true
	}

	return irBoolVar{
		name:  varName,
		desc:  varDesc,
		value: hvar,
		unit:  varUnit,
	}
}

func extractIntFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irIntVar {
	var hvar C.int // 4 bytes

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	return irIntVar{
		name:  varName,
		desc:  varDesc,
		value: int(hvar),
		unit:  varUnit,
	}
}

func extractBitfieldFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irBitfieldVar {
	// - SessionFlags
	// - CamCameraState
	// - EngineWarnings
	var hvar uint32 // 4 bytes

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	retVar := irBitfieldVar{
		name:   varName,
		desc:   varDesc,
		fields: make(map[string]bool),
		unit:   varUnit,
	}

	switch varName {
	case "SessionFlags":
		for bitmask, name := range irsdkFlags {
			retVar.fields[name] = bool(uint32(hvar)&uint32(bitmask) != 0)
		}
	case "CamCameraState":
		for bitmask, name := range irsdkCameraStates {
			retVar.fields[name] = bool(uint32(hvar)&uint32(bitmask) != 0)
		}
	case "EngineWarnings":
		for bitmask, name := range irsdkEngineWarnings {
			retVar.fields[name] = bool(uint32(hvar)&uint32(bitmask) != 0)
		}
	default:
		log.Println("Unknown bitField var:", varName)
	}

	return retVar
}

func extractFloatFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irFloatVar {
	var hvar C.float

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	return irFloatVar{
		name:  varName,
		desc:  varDesc,
		value: float32(hvar),
		unit:  varUnit,
	}
}

func extractDoubleFromVarHeader(header *utils.Irsdk_varHeader, data []byte) irDoubleVar {
	var hvar C.double

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	return irDoubleVar{
		name:  varName,
		desc:  varDesc,
		value: float64(hvar),
		unit:  varUnit,
	}
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
