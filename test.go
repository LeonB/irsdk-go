package main

import "C"
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
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

type TelemetryData struct {
	// bools
	DriverMarker                   bool
	IsOnTrack                      bool
	IsReplayPlaying                bool
	IsDiskLoggingEnabled           bool
	IsDiskLoggingActive            bool
	CarIdxOnPitRoad                bool
	OnPitRoad                      bool
	LapDeltaToBestLap_OK           bool
	LapDeltaToOptimalLap_OK        bool
	LapDeltaToSessionBestLap_OK    bool
	LapDeltaToSessionOptimalLap_OK bool
	LapDeltaToSessionLastlLap_OK   bool
	IsOnTrackCar                   bool
	IsInGarage                     bool
	ReplayPlaySlowMotion           bool

	// ints
	SessionNum                int
	SessionState              int
	SessionUniqueID           int
	SessionLapsRemain         int
	RadioTransmitCarIdx       int
	RadioTransmitRadioIdx     int
	RadioTransmitFrequencyIdx int
	ReplayFrameNum            int
	ReplayFrameNumEnd         int
	CarIdxLap                 int
	CarIdxTrackSurface        int
	CarIdxGear                int
	Gear                      int
	Lap                       int
	RaceLaps                  int
	LapBestLap                int
	CamCarIdx                 int
	CamCameraNumber           int
	CamGroupNumber            int
	ReplayPlaySpeed           int
	ReplaySessionNum          int

	// bitfields
	SessionFlags   map[string]bool
	CamCameraState map[string]bool
	EngineWarnings map[string]bool

	// floats
	FrameRate                       float32
	CpuUsageBG                      float32
	CarIdxLapDistPct                float32
	CarIdxSteer                     float32
	CarIdxRPM                       float32
	SteeringWheelAngle              float32
	Throttle                        float32
	Brake                           float32
	Clutch                          float32
	RPM                             float32
	LapDist                         float32
	LapDistPct                      float32
	LapBestLapTime                  float32
	LapLastLapTime                  float32
	LapCurrentLapTime               float32
	LapDeltaToBestLap               float32
	LapDeltaToBestLap_DD            float32
	LapDeltaToOptimalLap            float32
	LapDeltaToOptimalLap_DD         float32
	LapDeltaToSessionBestLap        float32
	LapDeltaToSessionBestLap_DD     float32
	LapDeltaToSessionOptimalLap     float32
	LapDeltaToSessionOptimalLap_DD  float32
	LapDeltaToSessionLastlLap       float32
	LapDeltaToSessionLastlLap_DD    float32
	LongAccel                       float32
	LatAccel                        float32
	VertAccel                       float32
	RollRate                        float32
	PitchRate                       float32
	YawRate                         float32
	Speed                           float32
	VelocityX                       float32
	VelocityY                       float32
	VelocityZ                       float32
	Yaw                             float32
	Pitch                           float32
	Roll                            float32
	PitRepairLeft                   float32
	PitOptRepairLeft                float32
	SteeringWheelTorque             float32
	SteeringWheelPctTorque          float32
	SteeringWheelPctTorqueSign      float32
	SteeringWheelPctTorqueSignStops float32
	SteeringWheelPctDamper          float32
	SteeringWheelAngleMax           float32
	ShiftIndicatorPct               float32
	ShiftPowerPct                   float32
	ShiftGrindRPM                   float32
	ThrottleRaw                     float32
	BrakeRaw                        float32
	SteeringWheelPeakForceNm        float32
	FuelLevel                       float32
	FuelLevelPct                    float32
	WaterTemp                       float32
	WaterLevel                      float32
	FuelPress                       float32
	OilTemp                         float32
	OilPress                        float32
	OilLevel                        float32
	Voltage                         float32
	ManifoldPress                   float32
	RRbrakeLinePress                float32
	RRcoldPressure                  float32
	RRtempCL                        float32
	RRtempCM                        float32
	RRtempCR                        float32
	RRwearL                         float32
	RRwearM                         float32
	RRwearR                         float32
	LRbrakeLinePress                float32
	LRcoldPressure                  float32
	LRtempCL                        float32
	LRtempCM                        float32
	LRtempCR                        float32
	LRwearL                         float32
	LRwearM                         float32
	LRwearR                         float32
	RFbrakeLinePress                float32
	RFcoldPressure                  float32
	RFtempCL                        float32
	RFtempCM                        float32
	RFtempCR                        float32
	RFwearL                         float32
	RFwearM                         float32
	RFwearR                         float32
	LFbrakeLinePress                float32
	LFcoldPressure                  float32
	LFtempCL                        float32
	LFtempCM                        float32
	LFtempCR                        float32
	LFwearL                         float32
	LFwearM                         float32
	LFwearR                         float32
	RRshockDefl                     float32
	RRshockVel                      float32
	LRshockDefl                     float32
	LRshockVel                      float32
	RFshockDefl                     float32
	RFshockVel                      float32
	LFshockDefl                     float32
	LFshockVel                      float32

	// Doubles
	SessionTime       float64
	SessionTimeRemain float64
	ReplaySessionTime float64
}

func (d *TelemetryData) AddIrCharVar(irVar *irCharVar) error {
	return nil
}

func (d *TelemetryData) AddIrBoolVar(irVar *irBoolVar) error {
	e := reflect.ValueOf(d).Elem() // Get reference to struct
	f := e.FieldByName(irVar.name) // Find struct field
	if f.Kind() == reflect.Bool {
		// A Value can be changed only if it is
		// addressable and was not obtained by
		// the use of unexported struct fields.
		if f.CanSet() {
			f.SetBool(irVar.value)
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Unknown %T: %v", irVar, irVar.name))
}

func (d *TelemetryData) AddIrIntVar(irVar *irIntVar) error {
	e := reflect.ValueOf(d).Elem() // Get reference to struct
	f := e.FieldByName(irVar.name) // Find struct field
	if f.Kind() == reflect.Int {
		if f.CanSet() {
			f.SetInt(int64(irVar.value))
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Unknown %T: %v", irVar, irVar.name))
}

func (d *TelemetryData) AddIrBitfieldVar(irVar *irBitfieldVar) error {
	e := reflect.ValueOf(d).Elem() // Get reference to struct
	f := e.FieldByName(irVar.name) // Find struct field
	if f.Kind() == reflect.Map {
		if f.CanSet() {
			for key, val := range irVar.fields {
				rKey := reflect.ValueOf(key)
				rVal := reflect.ValueOf(val)
				f.SetMapIndex(rKey, rVal)
			}
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Unknown %T: %v", irVar, irVar.name))
}

func (d *TelemetryData) AddIrFloatVar(irVar *irFloatVar) error {
	e := reflect.ValueOf(d).Elem() // Get reference to struct
	f := e.FieldByName(irVar.name) // Find struct field
	if f.Kind() == reflect.Float32 {
		if f.CanSet() {
			f.SetFloat(float64(irVar.value))
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Unknown %T: %v", irVar, irVar.name))
}

func (d *TelemetryData) AddIrDoubleVar(irVar *irDoubleVar) error {
	e := reflect.ValueOf(d).Elem() // Get reference to struct
	f := e.FieldByName(irVar.name) // Find struct field
	if f.Kind() == reflect.Float64 {
		if f.CanSet() {
			f.SetFloat(irVar.value)
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Unknown %T: %v", irVar, irVar.name))
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

var irsdkSessionStates = map[utils.Irsdk_SessionState]string{
	utils.Irsdk_StateInvalid:    "invalid",
	utils.Irsdk_StateGetInCar:   "getInCar",
	utils.Irsdk_StateWarmup:     "warmup",
	utils.Irsdk_StateParadeLaps: "paradeLaps",
	utils.Irsdk_StateRacing:     "racing",
	utils.Irsdk_StateCheckered:  "checkered",
	utils.Irsdk_StateCoolDown:   "coolDown",
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
			telemetryData := toTelemetryData(data)
			b, err := json.Marshal(telemetryData)
			if err != nil {
				fmt.Printf("Error: %s", err)
				return
			}
			fmt.Println(string(b))
		}

		utils.Irsdk_shutdown()
		break
	}

	return
}

func toTelemetryData(data []byte) *TelemetryData {
	telemetryData := newTelemetryData()
	numVars := utils.Irsdk_getNumVars()

	for i := 0; i <= numVars; i++ {
		varHeader := utils.Irsdk_getVarHeaderEntry(i)

		if varHeader == nil {
			continue
		}

		switch varHeader.Type {
		case utils.Irsdk_char:
			irVar := extractCharFromVarHeader(varHeader, data)
			err := telemetryData.AddIrCharVar(irVar)
			if err != nil {
				fmt.Println(err)
			}
		case utils.Irsdk_bool:
			irVar := extractBoolFromVarHeader(varHeader, data)
			err := telemetryData.AddIrBoolVar(irVar)
			if err != nil {
				fmt.Println(err)
			}
		case utils.Irsdk_int:
			irVar := extractIntFromVarHeader(varHeader, data)
			err := telemetryData.AddIrIntVar(irVar)
			if err != nil {
				fmt.Println(err)
			}
		case utils.Irsdk_bitField:
			irVar := extractBitfieldFromVarHeader(varHeader, data)
			err := telemetryData.AddIrBitfieldVar(irVar)
			if err != nil {
				fmt.Println(err)
			}
		case utils.Irsdk_float:
			irVar := extractFloatFromVarHeader(varHeader, data)
			err := telemetryData.AddIrFloatVar(irVar)
			if err != nil {
				fmt.Println(err)
			}
		case utils.Irsdk_double:
			irVar := extractDoubleFromVarHeader(varHeader, data)
			err := telemetryData.AddIrDoubleVar(irVar)
			if err != nil {
				fmt.Println(err)
			}
		default:
			log.Println("Unknown irsdk varType:", varHeader.Type)
		}
	}

	return telemetryData
}

func extractCharFromVarHeader(header *utils.Irsdk_varHeader, data []byte) *irCharVar {
	var hvar C.char // 1 byte

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	return &irCharVar{
		name:  varName,
		desc:  varDesc,
		value: byte(hvar),
		unit:  varUnit,
	}
}

func extractBoolFromVarHeader(header *utils.Irsdk_varHeader, data []byte) *irBoolVar {
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

	return &irBoolVar{
		name:  varName,
		desc:  varDesc,
		value: hvar,
		unit:  varUnit,
	}
}

func extractIntFromVarHeader(header *utils.Irsdk_varHeader, data []byte) *irIntVar {
	var hvar C.int // 4 bytes

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	return &irIntVar{
		name:  varName,
		desc:  varDesc,
		value: int(hvar),
		unit:  varUnit,
	}
}

func extractBitfieldFromVarHeader(header *utils.Irsdk_varHeader, data []byte) *irBitfieldVar {
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

	retVar := &irBitfieldVar{
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

func extractFloatFromVarHeader(header *utils.Irsdk_varHeader, data []byte) *irFloatVar {
	var hvar C.float

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	return &irFloatVar{
		name:  varName,
		desc:  varDesc,
		value: float32(hvar),
		unit:  varUnit,
	}
}

func extractDoubleFromVarHeader(header *utils.Irsdk_varHeader, data []byte) *irDoubleVar {
	var hvar C.double

	startByte := int(header.Offset)
	varLen := int(unsafe.Sizeof(hvar))
	endByte := startByte + varLen
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	buf := bytes.NewBuffer(data[startByte:endByte])
	binary.Read(buf, binary.LittleEndian, &hvar)

	return &irDoubleVar{
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

func newTelemetryData() *TelemetryData {
	return &TelemetryData{
		SessionFlags:   make(map[string]bool),
		CamCameraState: make(map[string]bool),
		EngineWarnings: make(map[string]bool),
	}
}
