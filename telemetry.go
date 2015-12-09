package irsdk

import (
	"bytes"
	"fmt"
	"log"
	"unsafe"
)

// Make own ctypes so I don't have to import "C"
// I don't know if this is correct
// Are there differences in 32/64 bit machines?
type Ctype_char int32
type Ctype_int int32
type Ctype_float float32
type Ctype_double float64

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

func (d *TelemetryData) addVarHeaderData(varHeader *VarHeader, data []byte) error {
	switch varHeader.Type {
	case CharType:
		irVar := d.extractCharFromVarHeader(varHeader, data)
		err := d.AddIrCharVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case BoolType:
		irVar := d.extractBoolFromVarHeader(varHeader, data)
		err := d.AddIrBoolVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case IntType:
		irVar := d.extractIntFromVarHeader(varHeader, data)
		err := d.AddIrIntVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case BitfieldType:
		irVar := d.extractBitfieldFromVarHeader(varHeader, data)
		err := d.AddIrBitfieldVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case FloatType:
		irVar := d.extractFloatFromVarHeader(varHeader, data)
		err := d.AddIrFloatVar(irVar)
		if err != nil {
			log.Println(err)
		}
	case DoubleType:
		irVar := d.extractDoubleFromVarHeader(varHeader, data)
		err := d.AddIrDoubleVar(irVar)
		if err != nil {
			log.Println(err)
		}
	default:
		log.Println("Unknown irsdk varType:", varHeader.Type)
	}

	return nil
}

// go:generate go run bin/telemetry_vars/telemetry_vars.go
func (d *TelemetryData) AddIrCharVar(irVar *irCharVar) error {
	if irVar == nil {
		return nil
	}

	return nil
}

func (d *TelemetryData) AddIrBoolVar(irVar *irBoolVar) error {
	if irVar == nil {
		return nil
	}

	switch irVar.name {
	case "CarIdxOnPitRoad":
		d.CarIdxOnPitRoad = irVar.value
	case "DriverMarker":
		d.DriverMarker = irVar.value
	case "IsDiskLoggingActive":
		d.IsDiskLoggingActive = irVar.value
	case "IsDiskLoggingEnabled":
		d.IsDiskLoggingEnabled = irVar.value
	case "IsInGarage":
		d.IsInGarage = irVar.value
	case "IsOnTrack":
		d.IsOnTrack = irVar.value
	case "IsOnTrackCar":
		d.IsOnTrackCar = irVar.value
	case "IsReplayPlaying":
		d.IsReplayPlaying = irVar.value
	case "LapDeltaToBestLap_OK":
		d.LapDeltaToBestLap_OK = irVar.value
	case "LapDeltaToOptimalLap_OK":
		d.LapDeltaToOptimalLap_OK = irVar.value
	case "LapDeltaToSessionBestLap_OK":
		d.LapDeltaToSessionBestLap_OK = irVar.value
	case "LapDeltaToSessionLastlLap_OK":
		d.LapDeltaToSessionLastlLap_OK = irVar.value
	case "LapDeltaToSessionOptimalLap_OK":
		d.LapDeltaToSessionOptimalLap_OK = irVar.value
	case "OnPitRoad":
		d.OnPitRoad = irVar.value
	case "ReplayPlaySlowMotion":
		d.ReplayPlaySlowMotion = irVar.value
	default:
		return fmt.Errorf("Unknown var: %+v", irVar)
	}

	return nil
}

func (d *TelemetryData) AddIrIntVar(irVar *irIntVar) error {
	if irVar == nil {
		return nil
	}

	switch irVar.name {
	case "CamCameraNumber":
		d.CamCameraNumber = irVar.value
	case "CamCarIdx":
		d.CamCarIdx = irVar.value
	case "CamGroupNumber":
		d.CamGroupNumber = irVar.value
	case "CarIdxClassPosition":
		d.CarIdxClassPosition = irVar.value
	case "CarIdxGear":
		d.CarIdxGear = irVar.value
	case "CarIdxLap":
		d.CarIdxLap = irVar.value
	case "CarIdxPosition":
		d.CarIdxPosition = irVar.value
	case "CarIdxTrackSurface":
		d.CarIdxTrackSurface = irVar.value
	case "DCDriversSoFar":
		d.DCDriversSoFar = irVar.value
	case "DCLapStatus":
		d.DCLapStatus = irVar.value
	case "DisplayUnits":
		d.DisplayUnits = irVar.value
	case "EnterExitReset":
		d.EnterExitReset = irVar.value
	case "Gear":
		d.Gear = irVar.value
	case "Lap":
		d.Lap = irVar.value
	case "LapBestLap":
		d.LapBestLap = irVar.value
	case "LapBestNLapLap":
		d.LapBestNLapLap = irVar.value
	case "LapLasNLapSeq":
		d.LapLasNLapSeq = irVar.value
	case "PlayerCarClassPosition":
		d.PlayerCarClassPosition = irVar.value
	case "PlayerCarPosition":
		d.PlayerCarPosition = irVar.value
	case "RaceLaps":
		d.RaceLaps = irVar.value
	case "RadioTransmitCarIdx":
		d.RadioTransmitCarIdx = irVar.value
	case "RadioTransmitFrequencyIdx":
		d.RadioTransmitFrequencyIdx = irVar.value
	case "RadioTransmitRadioIdx":
		d.RadioTransmitRadioIdx = irVar.value
	case "ReplayFrameNum":
		d.ReplayFrameNum = irVar.value
	case "ReplayFrameNumEnd":
		d.ReplayFrameNumEnd = irVar.value
	case "ReplayPlaySpeed":
		d.ReplayPlaySpeed = irVar.value
	case "ReplaySessionNum":
		d.ReplaySessionNum = irVar.value
	case "SessionLapsRemain":
		d.SessionLapsRemain = irVar.value
	case "SessionNum":
		d.SessionNum = irVar.value
	case "SessionState":
		d.SessionState = irVar.value
	case "SessionUniqueID":
		d.SessionUniqueID = irVar.value
	case "Skies":
		d.Skies = irVar.value
	case "WeatherType":
		d.WeatherType = irVar.value
	default:
		return fmt.Errorf("Unknown var: %+v", irVar)
	}

	return nil
}

func (d *TelemetryData) AddIrBitfieldVar(irVar *irBitfieldVar) error {
	if irVar == nil {
		return nil
	}

	switch irVar.name {
	case "CamCameraState":
		d.CamCameraState = irVar.fields
	case "EngineWarnings":
		d.EngineWarnings = irVar.fields
	case "PitSvFlags":
		d.PitSvFlags = irVar.fields
	case "SessionFlags":
		d.SessionFlags = irVar.fields
	default:
		return fmt.Errorf("Unknown var: %+v", irVar)
	}

	return nil
}

func (d *TelemetryData) AddIrFloatVar(irVar *irFloatVar) error {
	if irVar == nil {
		return nil
	}

	switch irVar.name {
	case "AirDensity":
		d.AirDensity = irVar.value
	case "AirPressure":
		d.AirPressure = irVar.value
	case "AirTemp":
		d.AirTemp = irVar.value
	case "Alt":
		d.Alt = irVar.value
	case "Brake":
		d.Brake = irVar.value
	case "BrakeRaw":
		d.BrakeRaw = irVar.value
	case "CarIdxEstTime":
		d.CarIdxEstTime = irVar.value
	case "CarIdxF2Time":
		d.CarIdxF2Time = irVar.value
	case "CarIdxLapDistPct":
		d.CarIdxLapDistPct = irVar.value
	case "CarIdxRPM":
		d.CarIdxRPM = irVar.value
	case "CarIdxSteer":
		d.CarIdxSteer = irVar.value
	case "CFSRrideHeight":
		d.CFSRrideHeight = irVar.value
	case "Clutch":
		d.Clutch = irVar.value
	case "CpuUsageBG":
		d.CpuUsageBG = irVar.value
	case "dcABS":
		d.DcABS = irVar.value
	case "dcBrakeBias":
		d.DcBrakeBias = irVar.value
	case "dcFuelMixture":
		d.DcFuelMixture = irVar.value
	case "dcThrottleShape":
		d.DcThrottleShape = irVar.value
	case "dcTractionControl":
		d.DcTractionControl = irVar.value
	case "FogLevel":
		d.FogLevel = irVar.value
	case "FrameRate":
		d.FrameRate = irVar.value
	case "FuelLevel":
		d.FuelLevel = irVar.value
	case "FuelLevelPct":
		d.FuelLevelPct = irVar.value
	case "FuelPress":
		d.FuelPress = irVar.value
	case "LapBestLapTime":
		d.LapBestLapTime = irVar.value
	case "LapBestNLapTime":
		d.LapBestNLapTime = irVar.value
	case "LapCurrentLapTime":
		d.LapCurrentLapTime = irVar.value
	case "LapDeltaToBestLap":
		d.LapDeltaToBestLap = irVar.value
	case "LapDeltaToBestLap_DD":
		d.LapDeltaToBestLap_DD = irVar.value
	case "LapDeltaToOptimalLap":
		d.LapDeltaToOptimalLap = irVar.value
	case "LapDeltaToOptimalLap_DD":
		d.LapDeltaToOptimalLap_DD = irVar.value
	case "LapDeltaToSessionBestLap":
		d.LapDeltaToSessionBestLap = irVar.value
	case "LapDeltaToSessionBestLap_DD":
		d.LapDeltaToSessionBestLap_DD = irVar.value
	case "LapDeltaToSessionLastlLap":
		d.LapDeltaToSessionLastlLap = irVar.value
	case "LapDeltaToSessionLastlLap_DD":
		d.LapDeltaToSessionLastlLap_DD = irVar.value
	case "LapDeltaToSessionOptimalLap":
		d.LapDeltaToSessionOptimalLap = irVar.value
	case "LapDeltaToSessionOptimalLap_DD":
		d.LapDeltaToSessionOptimalLap_DD = irVar.value
	case "LapDist":
		d.LapDist = irVar.value
	case "LapDistPct":
		d.LapDistPct = irVar.value
	case "LapLastLapTime":
		d.LapLastLapTime = irVar.value
	case "LapLastNLapTime":
		d.LapLastNLapTime = irVar.value
	case "LatAccel":
		d.LatAccel = irVar.value
	case "LFbrakeLinePress":
		d.LFbrakeLinePress = irVar.value
	case "LFcoldPressure":
		d.LFcoldPressure = irVar.value
	case "LFpressure":
		d.LFpressure = irVar.value
	case "LFrideHeight":
		d.LFrideHeight = irVar.value
	case "LFshockDefl":
		d.LFshockDefl = irVar.value
	case "LFshockVel":
		d.LFshockVel = irVar.value
	case "LFspeed":
		d.LFspeed = irVar.value
	case "LFtempCL":
		d.LFtempCL = irVar.value
	case "LFtempCM":
		d.LFtempCM = irVar.value
	case "LFtempCR":
		d.LFtempCR = irVar.value
	case "LFtempL":
		d.LFtempL = irVar.value
	case "LFtempM":
		d.LFtempM = irVar.value
	case "LFtempR":
		d.LFtempR = irVar.value
	case "LFwearL":
		d.LFwearL = irVar.value
	case "LFwearM":
		d.LFwearM = irVar.value
	case "LFwearR":
		d.LFwearR = irVar.value
	case "LongAccel":
		d.LongAccel = irVar.value
	case "LRbrakeLinePress":
		d.LRbrakeLinePress = irVar.value
	case "LRcoldPressure":
		d.LRcoldPressure = irVar.value
	case "LRpressure":
		d.LRpressure = irVar.value
	case "LRrideHeight":
		d.LRrideHeight = irVar.value
	case "LRshockDefl":
		d.LRshockDefl = irVar.value
	case "LRshockVel":
		d.LRshockVel = irVar.value
	case "LRspeed":
		d.LRspeed = irVar.value
	case "LRtempCL":
		d.LRtempCL = irVar.value
	case "LRtempCM":
		d.LRtempCM = irVar.value
	case "LRtempCR":
		d.LRtempCR = irVar.value
	case "LRtempL":
		d.LRtempL = irVar.value
	case "LRtempM":
		d.LRtempM = irVar.value
	case "LRtempR":
		d.LRtempR = irVar.value
	case "LRwearL":
		d.LRwearL = irVar.value
	case "LRwearM":
		d.LRwearM = irVar.value
	case "LRwearR":
		d.LRwearR = irVar.value
	case "ManifoldPress":
		d.ManifoldPress = irVar.value
	case "OilLevel":
		d.OilLevel = irVar.value
	case "OilPress":
		d.OilPress = irVar.value
	case "OilTemp":
		d.OilTemp = irVar.value
	case "Pitch":
		d.Pitch = irVar.value
	case "PitchRate":
		d.PitchRate = irVar.value
	case "PitOptRepairLeft":
		d.PitOptRepairLeft = irVar.value
	case "PitRepairLeft":
		d.PitRepairLeft = irVar.value
	case "PitSvFuel":
		d.PitSvFuel = irVar.value
	case "PitSvLFP":
		d.PitSvLFP = irVar.value
	case "PitSvLRP":
		d.PitSvLRP = irVar.value
	case "PitSvRFP":
		d.PitSvRFP = irVar.value
	case "PitSvRRP":
		d.PitSvRRP = irVar.value
	case "RelativeHumidity":
		d.RelativeHumidity = irVar.value
	case "RFbrakeLinePress":
		d.RFbrakeLinePress = irVar.value
	case "RFcoldPressure":
		d.RFcoldPressure = irVar.value
	case "RFpressure":
		d.RFpressure = irVar.value
	case "RFrideHeight":
		d.RFrideHeight = irVar.value
	case "RFshockDefl":
		d.RFshockDefl = irVar.value
	case "RFshockVel":
		d.RFshockVel = irVar.value
	case "RFspeed":
		d.RFspeed = irVar.value
	case "RFtempCL":
		d.RFtempCL = irVar.value
	case "RFtempCM":
		d.RFtempCM = irVar.value
	case "RFtempCR":
		d.RFtempCR = irVar.value
	case "RFtempL":
		d.RFtempL = irVar.value
	case "RFtempM":
		d.RFtempM = irVar.value
	case "RFtempR":
		d.RFtempR = irVar.value
	case "RFwearL":
		d.RFwearL = irVar.value
	case "RFwearM":
		d.RFwearM = irVar.value
	case "RFwearR":
		d.RFwearR = irVar.value
	case "Roll":
		d.Roll = irVar.value
	case "RollRate":
		d.RollRate = irVar.value
	case "RPM":
		d.RPM = irVar.value
	case "RRbrakeLinePress":
		d.RRbrakeLinePress = irVar.value
	case "RRcoldPressure":
		d.RRcoldPressure = irVar.value
	case "RRpressure":
		d.RRpressure = irVar.value
	case "RRrideHeight":
		d.RRrideHeight = irVar.value
	case "RRshockDefl":
		d.RRshockDefl = irVar.value
	case "RRshockVel":
		d.RRshockVel = irVar.value
	case "RRspeed":
		d.RRspeed = irVar.value
	case "RRtempCL":
		d.RRtempCL = irVar.value
	case "RRtempCM":
		d.RRtempCM = irVar.value
	case "RRtempCR":
		d.RRtempCR = irVar.value
	case "RRtempL":
		d.RRtempL = irVar.value
	case "RRtempM":
		d.RRtempM = irVar.value
	case "RRtempR":
		d.RRtempR = irVar.value
	case "RRwearL":
		d.RRwearL = irVar.value
	case "RRwearM":
		d.RRwearM = irVar.value
	case "RRwearR":
		d.RRwearR = irVar.value
	case "ShiftGrindRPM":
		d.ShiftGrindRPM = irVar.value
	case "ShiftIndicatorPct":
		d.ShiftIndicatorPct = irVar.value
	case "ShiftPowerPct":
		d.ShiftPowerPct = irVar.value
	case "Speed":
		d.Speed = irVar.value
	case "SteeringWheelAngle":
		d.SteeringWheelAngle = irVar.value
	case "SteeringWheelAngleMax":
		d.SteeringWheelAngleMax = irVar.value
	case "SteeringWheelPctDamper":
		d.SteeringWheelPctDamper = irVar.value
	case "SteeringWheelPctTorque":
		d.SteeringWheelPctTorque = irVar.value
	case "SteeringWheelPctTorqueSign":
		d.SteeringWheelPctTorqueSign = irVar.value
	case "SteeringWheelPctTorqueSignStops":
		d.SteeringWheelPctTorqueSignStops = irVar.value
	case "SteeringWheelPeakForceNm":
		d.SteeringWheelPeakForceNm = irVar.value
	case "SteeringWheelTorque":
		d.SteeringWheelTorque = irVar.value
	case "Throttle":
		d.Throttle = irVar.value
	case "ThrottleRaw":
		d.ThrottleRaw = irVar.value
	case "TrackTemp":
		d.TrackTemp = irVar.value
	case "VelocityX":
		d.VelocityX = irVar.value
	case "VelocityY":
		d.VelocityY = irVar.value
	case "VelocityZ":
		d.VelocityZ = irVar.value
	case "VertAccel":
		d.VertAccel = irVar.value
	case "Voltage":
		d.Voltage = irVar.value
	case "WaterLevel":
		d.WaterLevel = irVar.value
	case "WaterTemp":
		d.WaterTemp = irVar.value
	case "WindDir":
		d.WindDir = irVar.value
	case "WindVel":
		d.WindVel = irVar.value
	case "Yaw":
		d.Yaw = irVar.value
	case "YawRate":
		d.YawRate = irVar.value
	default:
		return fmt.Errorf("Unknown var: %+v", irVar)
	}

	return nil
}

func (d *TelemetryData) AddIrDoubleVar(irVar *irDoubleVar) error {
	if irVar == nil {
		return nil
	}

	switch irVar.name {
	case "Lat":
		d.Lat = irVar.value
	case "Lon":
		d.Lon = irVar.value
	case "ReplaySessionTime":
		d.ReplaySessionTime = irVar.value
	case "SessionTime":
		d.SessionTime = irVar.value
	case "SessionTimeRemain":
		d.SessionTimeRemain = irVar.value
	default:
		return fmt.Errorf("Unknown var: %+v", irVar)
	}

	return nil
}

var irsdkFlags = map[Flags]string{
	// global flags
	CheckeredFlag:     "Checkered",
	WhiteFlag:         "White",
	GreenFlag:         "Green",
	YellowFlag:        "Yellow",
	RedFlag:           "Red",
	BlueFlag:          "Blue",
	DebrisFlag:        "Debris",
	CrossedFlag:       "Crossed",
	YellowWavingFlag:  "YellowWaving",
	OneLapToGreenFlag: "OneLapToGreen",
	GreenHeldFlag:     "GreenHeld",
	TenToGoFlag:       "TenToGo",
	FiveToGoFlag:      "FiveToGo",
	RandomWavingFlag:  "RandomWaving",
	CautionFlag:       "Caution",
	CautionWavingFlag: "CautionWaving",

	// drivers black flags
	BlackFlag:      "Black",
	DisqualifyFlag: "Disqualify",
	ServicibleFlag: "Servicible", // car is allowed service (not a flag)
	FurledFlag:     "Furled",
	RepairFlag:     "Repair",

	// start lights
	StartHidden: "StartHidden",
	StartReady:  "StartReady",
	StartSet:    "StartSet",
	StartGo:     "StartGo",
}

var irsdkEngineWarnings = map[EngineWarnings]string{
	WaterTempWarning:    "WaterTempWarning",
	FuelPressureWarning: "FuelPressureWarning",
	OilPressureWarning:  "OilPressureWarning",
	EngineStalled:       "EngineStalled",
	PitSpeedLimiter:     "PitSpeedLimiter",
	RevLimiterActive:    "RevLimiterActive",
}

var irsdkCameraStates = map[CameraState]string{
	IsSessionScreen:       "IsSessionScreen",
	IsScenicActive:        "IsScencActive",
	CamToolActive:         "RamToolActive",
	UIHidden:              "UiHidden",
	UseAutoShotSelection:  "UseAutoShotSelection",
	UseTemporaryEdits:     "UseTemporaryEdits",
	UseKeyAcceleration:    "UseKeyAcceleration",
	UseKey10xAcceleration: "UseKey10xAcceleration",
	UseMouseAimMode:       "UseMouseAimMode",
}

var irsdkSessionStates = map[SessionState]string{
	StateInvalid:    "Invalid",
	StateGetInCar:   "GetInCar",
	StateWarmup:     "Warmup",
	StateParadeLaps: "ParadeLaps",
	StateRacing:     "Racing",
	StateCheckered:  "Checkered",
	StateCoolDown:   "CoolDown",
}

var irsdkPitSvFlags = map[PitSvFlag]string{
	LFTireChange: "LFTireChange",
	RFTireChange: "RFTireChange",
	LRTireChange: "LRTireChange",
	RRTireChange: "RRTireChange",

	FuelFill:          "FuelFill",
	WindshieldTearoff: "WindshieldTearoff",
	FastRepair:        "FastRepair",
}

func (d *TelemetryData) extractCharFromVarHeader(header *VarHeader, data []byte) *irCharVar {
	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*Ctype_char)(unsafe.Pointer(varPtr))

	return &irCharVar{
		name:  header.Name,
		desc:  header.Desc,
		value: byte(hvar),
		unit:  header.Unit,
	}
}

func (d *TelemetryData) extractBoolFromVarHeader(header *VarHeader, data []byte) *irBoolVar {
	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*bool)(unsafe.Pointer(varPtr))

	return &irBoolVar{
		name:  header.Name,
		desc:  header.Desc,
		value: hvar,
		unit:  header.Unit,
	}
}

func (d *TelemetryData) extractIntFromVarHeader(header *VarHeader, data []byte) *irIntVar {
	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*Ctype_int)(unsafe.Pointer(varPtr))

	return &irIntVar{
		name:  header.Name,
		desc:  header.Desc,
		value: int(hvar),
		unit:  header.Unit,
	}
}

func (d *TelemetryData) extractBitfieldFromVarHeader(header *VarHeader, data []byte) *irBitfieldVar {
	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*uint32)(unsafe.Pointer(varPtr))

	retVar := &irBitfieldVar{
		name:   header.Name,
		desc:   header.Desc,
		fields: make(map[string]bool),
		unit:   header.Unit,
	}

	switch header.Name {
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
		log.Println("Unknown bitField var:", header.Name)
	}

	return retVar
}

func (d *TelemetryData) extractFloatFromVarHeader(header *VarHeader, data []byte) *irFloatVar {
	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*Ctype_float)(unsafe.Pointer(varPtr))

	return &irFloatVar{
		name:  header.Name,
		desc:  header.Desc,
		value: float32(hvar),
		unit:  header.Unit,
	}
}

func (d *TelemetryData) extractDoubleFromVarHeader(header *VarHeader, data []byte) *irDoubleVar {
	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	offset := uintptr(header.Offset)
	varPtr := dataPtr + offset

	hvar := *(*Ctype_double)(unsafe.Pointer(varPtr))

	return &irDoubleVar{
		name:  header.Name,
		desc:  header.Desc,
		value: float64(hvar),
		unit:  header.Unit,
	}
}

func NewTelemetryData() *TelemetryData {
	return &TelemetryData{
		// fieldCache:     make(map[string]*reflect.Value),
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
