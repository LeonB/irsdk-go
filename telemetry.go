package irsdk

import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"unsafe"

	utils "github.com/leonb/irsdk-go/utils"
)

var telemetryData = NewTelemetryData()

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
	fieldCache map[string]*reflect.Value

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
	DisplayUnits              int
	PlayerCarPosition         int
	PlayerCarClassPosition    int
	CarIdxPosition            int
	CarIdxClassPosition       int
	LapLasNLapSeq             int
	LapBestNLapLap            int
	EnterExitReset            int
	DCLapStatus               int
	DCDriversSoFar            int

	// Only used in disk based telemetry
	WeatherType int
	Skies       int

	// bitfields
	SessionFlags   map[string]bool
	CamCameraState map[string]bool
	EngineWarnings map[string]bool

	// Only used in disk based telemetry data
	PitSvFlags map[string]bool

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
	CarIdxF2Time                    float32
	CarIdxEstTime                   float32
	LapLastNLapTime                 float32
	brakeLinePresse                 float32
	DcBrakeBias                     float32
	LapBestNLapTime                 float32

	// Only used in disk based telemetry
	Alt               float32
	TrackTemp         float32
	AirTemp           float32
	AirDensity        float32
	AirPressure       float32
	WindVel           float32
	WindDir           float32
	RelativeHumidity  float32
	LRtempL           float32
	LRtempM           float32
	LRtempR           float32
	RFspeed           float32
	RFpressure        float32
	DcABS             float32
	DcThrottleShape   float32
	DcFuelMixture     float32
	RRspeed           float32
	RRpressure        float32
	RRtempL           float32
	RRtempM           float32
	RRtempR           float32
	LRspeed           float32
	LRpressure        float32
	RFtempL           float32
	RFtempM           float32
	RFtempR           float32
	LFspeed           float32
	LFpressure        float32
	LFtempL           float32
	LFtempM           float32
	LFtempR           float32
	LFrideHeight      float32
	RFrideHeight      float32
	LRrideHeight      float32
	RRrideHeight      float32
	CFSRrideHeight    float32
	PitSvLRP          float32
	PitSvRRP          float32
	FogLevel          float32
	DcTractionControl float32
	PitSvLFP          float32
	PitSvRFP          float32
	PitSvFuel         float32

	// Doubles
	SessionTime       float64
	SessionTimeRemain float64
	ReplaySessionTime float64

	// Only used in disk based telemetry
	Lat float64
	Lon float64
}

func (d *TelemetryData) addVarHeaderData(varHeader *utils.VarHeader, data []byte) error {
	switch varHeader.Type {
	case utils.CharType:
		irVar := extractCharFromVarHeader(varHeader, data)
		err := d.AddIrCharVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case utils.BoolType:
		irVar := extractBoolFromVarHeader(varHeader, data)
		err := d.AddIrBoolVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case utils.IntType:
		irVar := extractIntFromVarHeader(varHeader, data)
		err := d.AddIrIntVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case utils.BitfieldType:
		irVar := extractBitfieldFromVarHeader(varHeader, data)
		err := d.AddIrBitfieldVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case utils.FloatType:
		irVar := extractFloatFromVarHeader(varHeader, data)
		err := d.AddIrFloatVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case utils.DoubleType:
		irVar := extractDoubleFromVarHeader(varHeader, data)
		err := d.AddIrDoubleVar(irVar)
		if err != nil {
			log.Println(err)
		}
	default:
		log.Println("Unknown irsdk varType:", varHeader.Type)
	}

	return nil
}

func (d *TelemetryData) fieldByName(varName string, kind reflect.Kind) (*reflect.Value, error) {
	if f, ok := d.fieldCache[varName]; ok {
		return f, nil
	}

	// Uppercase string because all are stored as public variables
	varNameUp := ucFirst(varName)

	e := reflect.ValueOf(d).Elem() // Get reference to struct
	f := e.FieldByName(varNameUp)  // Find struct field
	if f.Kind() == kind {
		if f.CanSet() {
			d.fieldCache[varName] = &f
			return &f, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Unknown %v/%v: %v", kind, f.Kind(), varNameUp))
}

func (d *TelemetryData) AddIrCharVar(irVar *irCharVar) error {
	return nil
}

func (d *TelemetryData) AddIrBoolVar(irVar *irBoolVar) error {
	if irVar == nil {
		return nil
	}

	f, err := d.fieldByName(irVar.name, reflect.Bool)
	if err != nil {
		return err
	}

	f.SetBool(irVar.value)
	return nil
}

func (d *TelemetryData) AddIrIntVar(irVar *irIntVar) error {
	if irVar == nil {
		return nil
	}

	f, err := d.fieldByName(irVar.name, reflect.Int)
	if err != nil {
		return err
	}

	f.SetInt(int64(irVar.value))
	return nil
}

func (d *TelemetryData) AddIrBitfieldVar(irVar *irBitfieldVar) error {
	if irVar == nil {
		return nil
	}

	f, err := d.fieldByName(irVar.name, reflect.Map)
	if err != nil {
		return err
	}

	for key, val := range irVar.fields {
		rKey := reflect.ValueOf(key)
		rVal := reflect.ValueOf(val)
		f.SetMapIndex(rKey, rVal)
	}
	return nil
}

func (d *TelemetryData) AddIrFloatVar(irVar *irFloatVar) error {
	if irVar == nil {
		return nil
	}

	f, err := d.fieldByName(irVar.name, reflect.Float32)
	if err != nil {
		return err
	}

	f.SetFloat(float64(irVar.value))
	return nil
}

func (d *TelemetryData) AddIrDoubleVar(irVar *irDoubleVar) error {
	if irVar == nil {
		return nil
	}

	f, err := d.fieldByName(irVar.name, reflect.Float64)
	if err != nil {
		return err
	}

	f.SetFloat(irVar.value)
	return nil
}

var irsdkFlags = map[utils.Flags]string{
	// global flags
	utils.CheckeredFlag:     "Checkered",
	utils.WhiteFlag:         "White",
	utils.GreenFlag:         "Green",
	utils.YellowFlag:        "Yellow",
	utils.RedFlag:           "Red",
	utils.BlueFlag:          "Blue",
	utils.DebrisFlag:        "Debris",
	utils.CrossedFlag:       "Crossed",
	utils.YellowWavingFlag:  "YellowWaving",
	utils.OneLapToGreenFlag: "OneLapToGreen",
	utils.GreenHeldFlag:     "GreenHeld",
	utils.TenToGoFlag:       "TenToGo",
	utils.FiveToGoFlag:      "FiveToGo",
	utils.RandomWavingFlag:  "RandomWaving",
	utils.CautionFlag:       "Caution",
	utils.CautionWavingFlag: "CautionWaving",

	// drivers black flags
	utils.BlackFlag:      "Black",
	utils.DisqualifyFlag: "Disqualify",
	utils.ServicibleFlag: "Servicible", // car is allowed service (not a flag)
	utils.FurledFlag:     "Furled",
	utils.RepairFlag:     "Repair",

	// start lights
	utils.StartHidden: "StartHidden",
	utils.StartReady:  "StartReady",
	utils.StartSet:    "StartSet",
	utils.StartGo:     "StartGo",
}

var irsdkEngineWarnings = map[utils.EngineWarnings]string{
	utils.WaterTempWarning:    "WaterTempWarning",
	utils.FuelPressureWarning: "FuelPressureWarning",
	utils.OilPressureWarning:  "OilPressureWarning",
	utils.EngineStalled:       "EngineStalled",
	utils.PitSpeedLimiter:     "PitSpeedLimiter",
	utils.RevLimiterActive:    "RevLimiterActive",
}

var irsdkCameraStates = map[utils.CameraState]string{
	utils.IsSessionScreen:       "IsSessionScreen",
	utils.IsScenicActive:        "IsScencActive",
	utils.CamToolActive:         "RamToolActive",
	utils.UIHidden:              "UiHidden",
	utils.UseAutoShotSelection:  "UseAutoShotSelection",
	utils.UseTemporaryEdits:     "UseTemporaryEdits",
	utils.UseKeyAcceleration:    "UseKeyAcceleration",
	utils.UseKey10xAcceleration: "UseKey10xAcceleration",
	utils.UseMouseAimMode:       "UseMouseAimMode",
}

var irsdkSessionStates = map[utils.SessionState]string{
	utils.StateInvalid:    "Invalid",
	utils.StateGetInCar:   "GetInCar",
	utils.StateWarmup:     "Warmup",
	utils.StateParadeLaps: "ParadeLaps",
	utils.StateRacing:     "Racing",
	utils.StateCheckered:  "Checkered",
	utils.StateCoolDown:   "CoolDown",
}

var irsdkPitSvFlags = map[utils.PitSvFlag]string{
	utils.LFTireChange: "LFTireChange",
	utils.RFTireChange: "RFTireChange",
	utils.LRTireChange: "LRTireChange",
	utils.RRTireChange: "RRTireChange",

	utils.FuelFill:          "FuelFill",
	utils.WindshieldTearoff: "WindshieldTearoff",
	utils.FastRepair:        "FastRepair",
}

// @TODO: should this accept an io.Reader?
func (c *Connection) BytesToTelemetryStruct(data []byte) (*TelemetryData, error) {
	// Create an new struct in the same memory location so reflect values can be
	// cached
	td := NewTelemetryData()
	td.fieldCache = telemetryData.fieldCache
	*telemetryData = *td
	numVars := c.sdk.GetNumVars()

	for i := 0; i <= numVars; i++ {
		varHeader, err := c.sdk.GetVarHeaderEntry(i)
		if err != nil {
			continue
		}

		if varHeader == nil {
			continue
		}

		telemetryData.addVarHeaderData(varHeader, data)
	}

	return telemetryData, nil
}

// @TODO: should this accept an io.Reader?
// @TODO: this shouldn't be on the connection because it can also be used by
// disk based telemetry (.ibt)
func (c *Connection) BytesToTelemetryStructFiltered(data []byte, fields []string) *TelemetryData {
	// Create an new struct in the same memory location so reflect values can be
	// cached
	td := NewTelemetryData()
	td.fieldCache = telemetryData.fieldCache
	*telemetryData = *td
	numVars := c.sdk.GetNumVars()

	for i := 0; i <= numVars; i++ {
		varHeader, err := c.sdk.GetVarHeaderEntry(i)
		if err != nil {
			continue
		}

		if varHeader == nil {
			continue
		}

		if fields == nil || len(fields) == 0 {
			// fields is empty: add everything
			telemetryData.addVarHeaderData(varHeader, data)
			continue
		}

		varName := utils.CToGoString(varHeader.Name[:])
		found := false

		for _, v := range fields {
			if v == varName {
				// Found varName in fields, skip looping through fields
				found = true
				break
			}
		}

		if found == false {
			// var not in fieds: skip varHeader
			continue
		}

		telemetryData.addVarHeaderData(varHeader, data)
	}

	return telemetryData
}

func extractCharFromVarHeader(header *utils.VarHeader, data []byte) *irCharVar {
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*C.char)(unsafe.Pointer(varPtr))

	return &irCharVar{
		name:  varName,
		desc:  varDesc,
		value: byte(hvar),
		unit:  varUnit,
	}
}

func extractBoolFromVarHeader(header *utils.VarHeader, data []byte) *irBoolVar {
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*bool)(unsafe.Pointer(varPtr))

	return &irBoolVar{
		name:  varName,
		desc:  varDesc,
		value: hvar,
		unit:  varUnit,
	}
}

func extractIntFromVarHeader(header *utils.VarHeader, data []byte) *irIntVar {
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*C.int)(unsafe.Pointer(varPtr))

	return &irIntVar{
		name:  varName,
		desc:  varDesc,
		value: int(hvar),
		unit:  varUnit,
	}
}

func extractBitfieldFromVarHeader(header *utils.VarHeader, data []byte) *irBitfieldVar {
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*uint32)(unsafe.Pointer(varPtr))

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
	case "PitSvFlags":
		for bitmask, name := range irsdkPitSvFlags {
			retVar.fields[name] = bool(uint32(hvar)&uint32(bitmask) != 0)
		}
	default:
		log.Println("Unknown bitField var:", varName)
	}

	return retVar
}

func extractFloatFromVarHeader(header *utils.VarHeader, data []byte) *irFloatVar {
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*C.float)(unsafe.Pointer(varPtr))

	return &irFloatVar{
		name:  varName,
		desc:  varDesc,
		value: float32(hvar),
		unit:  varUnit,
	}
}

func extractDoubleFromVarHeader(header *utils.VarHeader, data []byte) *irDoubleVar {
	varName := utils.CToGoString(header.Name[:])
	varDesc := utils.CToGoString(header.Desc[:])
	varUnit := utils.CToGoString(header.Unit[:])

	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*C.double)(unsafe.Pointer(varPtr))

	return &irDoubleVar{
		name:  varName,
		desc:  varDesc,
		value: float64(hvar),
		unit:  varUnit,
	}
}

func NewTelemetryData() *TelemetryData {
	return &TelemetryData{
		fieldCache:     make(map[string]*reflect.Value),
		SessionFlags:   make(map[string]bool),
		CamCameraState: make(map[string]bool),
		EngineWarnings: make(map[string]bool),
		PitSvFlags:     make(map[string]bool),
	}
}

func ucFirst(s string) string {
	b := []byte(s)
	b[0] = bytes.ToUpper(b[0:1])[0]
	return string(b)
	// r, n := utf8.DecodeRuneInString(s)
	// return string(unicode.ToUpper(r)) + s[n:]
}
