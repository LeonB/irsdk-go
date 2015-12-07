package irsdk

/*
 The IRSDK is a simple api that lets clients access telemetry data from the
 iRacing simulator. It is broken down into several parts:
 - Live data
   Live data is output from the sim into a shared memory mapped file.  Any
   application can open this memory mapped file and read the telemetry data
   out.  The format of this data was laid out in such a way that it should be
   possible to access from any language that can open a windows memory mapped
   file, without needing an external api.
   There are two different types of data that the telemetry outputs,
   sessionInfo and variables:

   Session info is for data that only needs to be updated every once in a
   while.  This data is output as a YAML formatted string.
   Variables, on the other hand, are output at a rate of 60 times a second.
   The varHeader struct defines each variable that the sim will output, while
   the varData struct gives details about the current line buffer that the vars
   are being written into.  Each variable is packed into a binary array with
   an offset and length stored in the varHeader.  The number of variables
   available can change depending on the car or session loaded.  But once the
   sim is running the variable list is locked down and will not change during a
   session.
   The sim writes a new line of variables every 16 ms, and then signals any
   listeners in order to wake them up to read the data.  Because the sim has no
   way of knowing when a listener is done reading the data, we triple buffer
   it in order to give all the clients enough time to read the data out.  This
   gives you a minimum of 16 ms to read the data out and process it.  So it is
   best to copy the data out before processing it.  You can use the function
   irsdk_waitForDataReady() to both wait for new data and copy the data to a
   local buffer.
 - Logged data
   Detailed information about the local drivers car can be logged to disk in
   the form of an ibt binary file.  This logging is enabled in the sim by
   typing alt-L at any time.  The ibt file format directly mirrors the format
   of the live data.
   It is stored as an irsdk_header followed immediately by an irsdk_diskSubHeader.
   After that the offsets in the irsdk_header point to the sessionInfo string,
   the varHeader, and the varBuffer.
 - Remote Conrol
   You can control the camera selections and playback of a replay tape, from
   any external application by sending a windows message with the
   irsdk_broadcastMsg() function.
*/

// Constant Definitions

type StatusField int32
type time_t int32

const (
	StatusConnected StatusField = 1
)

type VarType int32

// Stringer method
func (v VarType) String() string {
	switch v {
	case CharType:
		return "char"
	case BoolType:
		return "bool"
	case IntType:
		return "int"
	case BitfieldType:
		return "bitfield"
	case FloatType:
		return "float"
	case DoubleType:
		return "double"
	default:
		return "unkown"
	}
}

const (
	// 1 byte
	CharType VarType = iota
	BoolType VarType = iota

	// 4 bytes
	IntType      VarType = iota
	BitfieldType VarType = iota
	FloatType    VarType = iota

	// 8 bytes
	DoubleType VarType = iota

	// index, don't use
	ETCount = iota
)

// const doesn't work for some reason
var VarTypeBytes = [ETCount]VarType{
	1, // CharType
	1, // BoolType

	4, // IntType
	4, // BitfieldType
	4, // FloatType

	8, // DoubleType
}

// bit fields
type EngineWarnings int32

const (
	WaterTempWarning    EngineWarnings = 0x01
	FuelPressureWarning EngineWarnings = 0x02
	OilPressureWarning  EngineWarnings = 0x04
	EngineStalled       EngineWarnings = 0x08
	PitSpeedLimiter     EngineWarnings = 0x10
	RevLimiterActive    EngineWarnings = 0x20
)

type Flags uint32

const (
	// global flags
	CheckeredFlag     Flags = 0x00000001
	WhiteFlag         Flags = 0x00000002
	GreenFlag         Flags = 0x00000004
	YellowFlag        Flags = 0x00000008
	RedFlag           Flags = 0x00000010
	BlueFlag          Flags = 0x00000020
	DebrisFlag        Flags = 0x00000040
	CrossedFlag       Flags = 0x00000080
	YellowWavingFlag  Flags = 0x00000100
	OneLapToGreenFlag Flags = 0x00000200
	GreenHeldFlag     Flags = 0x00000400
	TenToGoFlag       Flags = 0x00000800
	FiveToGoFlag      Flags = 0x00001000
	RandomWavingFlag  Flags = 0x00002000
	CautionFlag       Flags = 0x00004000
	CautionWavingFlag Flags = 0x00008000

	// drivers black flags
	BlackFlag      Flags = 0x00010000
	DisqualifyFlag Flags = 0x00020000
	ServicibleFlag Flags = 0x00040000 // car is allowed service (not a flag)
	FurledFlag     Flags = 0x00080000
	RepairFlag     Flags = 0x00100000

	// start lights
	StartHidden Flags = 0x10000000
	StartReady  Flags = 0x20000000
	StartSet    Flags = 0x40000000
	StartGo     Flags = 0x80000000
)

type TrkLoc int32

// status
const (
	NotInWorld     TrkLoc = iota - 1
	OffTrack       TrkLoc = iota - 1
	InPitStall     TrkLoc = iota - 1
	AproachingPits TrkLoc = iota - 1
	OnTrac         TrkLoc = iota - 1
)

type SessionState int32

const (
	StateInvalid    SessionState = iota
	StateGetInCar   SessionState = iota
	StateWarmup     SessionState = iota
	StateParadeLaps SessionState = iota
	StateRacing     SessionState = iota
	StateCheckered  SessionState = iota
	StateCoolDown   SessionState = iota
)

type CameraState int32

const (
	IsSessionScreen CameraState = 0x0001 // the camera tool can only be activated if viewing the session screen (out of car)
	IsScenicActive  CameraState = 0x0002 // the scenic camera is active (no focus car)

	//these can be changed with a broadcast message
	CamToolActive         CameraState = 0x0004
	UIHidden              CameraState = 0x0008
	UseAutoShotSelection  CameraState = 0x0010
	UseTemporaryEdits     CameraState = 0x0020
	UseKeyAcceleration    CameraState = 0x0040
	UseKey10xAcceleration CameraState = 0x0080
	UseMouseAimMode       CameraState = 0x0100
)

type PitSvFlag int32

const (
	LFTireChange PitSvFlag = 0x0001
	RFTireChange PitSvFlag = 0x0002
	LRTireChange PitSvFlag = 0x0004
	RRTireChange PitSvFlag = 0x0008

	FuelFill          PitSvFlag = 0x0010
	WindshieldTearoff PitSvFlag = 0x0020
	FastRepair        PitSvFlag = 0x0040
)

//----
//

type VarHeaderRaw struct {
	Type   VarType // VarType
	Offset int32   `json:"-"` // offset fron start of buffer row
	Count  int32   `json:"-"` // number of entrys (array)
	// so length in bytes would be VarTypeBytes[type] * count

	Pad [1]int32 `json:"-"` // (16 byte align)

	Name [MAX_STRING]byte
	Desc [MAX_DESC]byte
	Unit [MAX_STRING]byte // something like "kg/m^2"
}

type VarHeader struct {
	Type   VarType // VarType
	Offset int32   `json:"-"` // offset fron start of buffer row
	Count  int32   `json:"-"` // number of entrys (array)
	// so length in bytes would be VarTypeBytes[type] * count

	Name string
	Desc string
	Unit string // something like "kg/m^2"
}

type VarBuf struct {
	TickCount int32    // used to detect changes in data
	BufOffset int32    // offset from header
	Pad       [2]int32 // (16 byte align)
}

type Header struct {
	Ver      int32       // api version 1 for now
	Status   StatusField // bitfield using StatusField
	TickRate int32       // ticks per second (60 or 360 etc)

	// session information, updated periodicaly
	SessionInfoUpdate int32 // Incremented when session info changes
	SessionInfoLen    int32 // Length in bytes of session info string
	SessionInfoOffset int32 // Session info, encoded in YAML format

	// State data, output at tickRate
	NumVars         int32 // length of array pointed to by varHeaderOffset
	VarHeaderOffset int32 // offset to VarHeader[numVars] array, Describes the variables recieved in varBuf

	NumBuf int32    // <= MAX_BUFS (3 for now)
	BufLen int32    // length in bytes for one line
	Pad1   [2]int32 // (16 byte align)
	VarBuf [MAX_BUFS]VarBuf
}

// sub header used when writing telemetry to disk
type DiskSubHeader struct {
	sessionStartDate   time_t
	sessionStartTime   float64
	sessionEndTime     float64
	sessionLapCount    int32
	sessionRecordCount int32
}

func (header *Header) GetLatestVarBufN() int {
	latest := 0
	for i := 0; i < int(header.NumBuf); i++ {
		if header.VarBuf[latest].TickCount < header.VarBuf[i].TickCount {
			latest = i
		}
	}

	return latest
}

//----
// Remote controll the sim by sending these windows messages
// camera and replay commands only work when you are out of your car,
// pit commands only work when in your car

type BroadcastMsg uint16

const (
	BroadcastCamSwitchPos          BroadcastMsg = 0    // car position, group, camera
	BroadcastCamSwitchNum          BroadcastMsg = iota // driver #, group, camera
	BroadcastCamSetState           BroadcastMsg = iota // CameraState, unused, unused
	BroadcastReplaySetPlaySpeed    BroadcastMsg = iota // speed, slowMotion, unused
	BroadcastReplaySetPlayPosition BroadcastMsg = iota // RpyPosMode, Frame Number (high, low)
	BroadcastReplaySearch          BroadcastMsg = iota // RpySrchMode, unused, unused
	BroadcastReplaySetState        BroadcastMsg = iota // RpyStateMode, unused, unused
	BroadcastReloadTextures        BroadcastMsg = iota // ReloadTexturesMode, carIdx, unused
	BroadcastChatComand            BroadcastMsg = iota // ChatCommandMode, subCommand, unused
	BroadcastPitCommand            BroadcastMsg = iota // PitCommandMode, parameter
	BroadcastTelemCommand          BroadcastMsg = iota // TelemCommandMode, unused, unused
	BroadcastLast                  BroadcastMsg = iota // unused placeholder
)

type ChatCommandMode int32

const (
	ChatCommand_Macro     ChatCommandMode = 0    // pass in a number from 1-15 representing the chat macro to launch
	ChatCommand_BeginChat ChatCommandMode = iota // Open up a new chat window
	ChatCommand_Reply     ChatCommandMode = iota // Reply to last private chat
	ChatCommand_Cancel    ChatCommandMode = iota // Close chat window
)

type PitCommandMode int32

const (
	PitCommand_Clear      PitCommandMode = 0    // Clear all pit checkboxes
	PitCommand_WS         PitCommandMode = iota // Clean the winshield, using one tear off
	PitCommand_Fuel       PitCommandMode = iota // Add fuel, optionally specify the amount to add in liters or pass '0' to use existing amount
	PitCommand_LF         PitCommandMode = iota // Change the left front tire, optionally specifying the pressure in KPa or pass '0' to use existing pressure
	PitCommand_RF         PitCommandMode = iota // right front
	PitCommand_LR         PitCommandMode = iota // left rear
	PitCommand_RR         PitCommandMode = iota // right rear
	PitCommand_ClearTires PitCommandMode = iota // Clear tire pit checkboxes
)

type TelemCommandMode int32

const (
	TelemCommand_Stop    TelemCommandMode = 0    // Turn telemetry recording off
	TelemCommand_Start   TelemCommandMode = iota // Turn telemetry recording on
	TelemCommand_Restart TelemCommandMode = iota // Write current file to disk and start a new one
)

type RpyStateMode int32

const (
	RpyState_EraseTape RpyStateMode = 0    // clear any data in the replay tape
	RpyState_Last      RpyStateMode = iota // unused place holder
)

type ReloadTexturesMode int32

const (
	ReloadTextures_All    ReloadTexturesMode = 0    // reload all textuers
	ReloadTextures_CarIdx ReloadTexturesMode = iota // reload only textures for the specific carIdx
)

// Search replay tape for events
type RpySrchMode int32

const (
	RpySrch_ToStart      RpySrchMode = 0
	RpySrch_ToEnd        RpySrchMode = iota
	RpySrch_PrevSession  RpySrchMode = iota
	RpySrch_NextSession  RpySrchMode = iota
	RpySrch_PrevLap      RpySrchMode = iota
	RpySrch_NextLap      RpySrchMode = iota
	RpySrch_PrevFrame    RpySrchMode = iota
	RpySrch_NextFrame    RpySrchMode = iota
	RpySrch_PrevIncident RpySrchMode = iota
	RpySrch_NextIncident RpySrchMode = iota
	RpySrch_Last         RpySrchMode = iota // unused placeholder
)

type RpyPosMode int32

const (
	RpyPos_Begin   RpyPosMode = 0
	RpyPos_Current RpyPosMode = iota
	RpyPos_End     RpyPosMode = iota
	RpyPos_Last    RpyPosMode = iota // unused placeholder
)

// BroadcastCamSwitchPos or BroadcastCamSwitchNum camera focus defines
// pass these in for the first parameter to select the 'focus at' types in the camera system.
type CsMode int32

const (
	CsFocusAtIncident CsMode = -3
	CsFocusAtLeader   CsMode = -2
	CsFocusAtExiting  CsMode = -1
	// ctFocusAtDriver + car number...
	CsFocusAtDriver CsMode = 0
)
