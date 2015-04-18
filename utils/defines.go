// +build windows

package utils

import "C"

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

type irsdk_StatusField int

const (
	irsdk_stConnected irsdk_StatusField = 1
)

type irsdk_VarType C.int

const (
	// 1 byte
	Irsdk_char irsdk_VarType = iota
	Irsdk_bool irsdk_VarType = iota

	// 4 bytes
	Irsdk_int      irsdk_VarType = iota
	Irsdk_bitField irsdk_VarType = iota
	Irsdk_float    irsdk_VarType = iota

	// 8 bytes
	Irsdk_double irsdk_VarType = iota

	// index, don't use
	Irsdk_ETCount = iota
)

// const doesn't work for some reason
var irsdk_VarTypeBytes = [Irsdk_ETCount]int{
	1, // irsdk_char
	1, // irsdk_bool

	4, // irsdk_int
	4, // irsdk_bitField
	4, // irsdk_float

	8, // irsdk_double
}

// bit fields
type Irsdk_EngineWarnings int

const (
	Irsdk_waterTempWarning    Irsdk_EngineWarnings = 0x01
	Irsdk_fuelPressureWarning Irsdk_EngineWarnings = 0x02
	Irsdk_oilPressureWarning  Irsdk_EngineWarnings = 0x04
	Irsdk_engineStalled       Irsdk_EngineWarnings = 0x08
	Irsdk_pitSpeedLimiter     Irsdk_EngineWarnings = 0x10
	Irsdk_revLimiterActive    Irsdk_EngineWarnings = 0x20
)

type Irsdk_Flags uint32

const (
	// global flags
	Irsdk_checkered     Irsdk_Flags = 0x00000001
	Irsdk_white         Irsdk_Flags = 0x00000002
	Irsdk_green         Irsdk_Flags = 0x00000004
	Irsdk_yellow        Irsdk_Flags = 0x00000008
	Irsdk_red           Irsdk_Flags = 0x00000010
	Irsdk_blue          Irsdk_Flags = 0x00000020
	Irsdk_debris        Irsdk_Flags = 0x00000040
	Irsdk_crossed       Irsdk_Flags = 0x00000080
	Irsdk_yellowWaving  Irsdk_Flags = 0x00000100
	Irsdk_oneLapToGreen Irsdk_Flags = 0x00000200
	Irsdk_greenHeld     Irsdk_Flags = 0x00000400
	Irsdk_tenToGo       Irsdk_Flags = 0x00000800
	Irsdk_fiveToGo      Irsdk_Flags = 0x00001000
	Irsdk_randomWaving  Irsdk_Flags = 0x00002000
	Irsdk_caution       Irsdk_Flags = 0x00004000
	Irsdk_cautionWaving Irsdk_Flags = 0x00008000

	// drivers black flags
	Irsdk_black      Irsdk_Flags = 0x00010000
	Irsdk_disqualify Irsdk_Flags = 0x00020000
	Irsdk_servicible Irsdk_Flags = 0x00040000 // car is allowed service (not a flag)
	Irsdk_furled     Irsdk_Flags = 0x00080000
	Irsdk_repair     Irsdk_Flags = 0x00100000

	// start lights
	Irsdk_startHidden Irsdk_Flags = 0x10000000
	Irsdk_startReady  Irsdk_Flags = 0x20000000
	Irsdk_startSet    Irsdk_Flags = 0x40000000
	Irsdk_startGo     Irsdk_Flags = 0x80000000
)

type irsdk_TrkLoc int

// status
const (
	irsdk_NotInWorld     irsdk_TrkLoc = iota - 1
	irsdk_OffTrack       irsdk_TrkLoc = iota - 1
	irsdk_InPitStall     irsdk_TrkLoc = iota - 1
	irsdk_AproachingPits irsdk_TrkLoc = iota - 1
	irsdk_OnTrac         irsdk_TrkLoc = iota - 1
)

type Irsdk_SessionState int

const (
	Irsdk_StateInvalid    Irsdk_SessionState = iota
	Irsdk_StateGetInCar   Irsdk_SessionState = iota
	Irsdk_StateWarmup     Irsdk_SessionState = iota
	Irsdk_StateParadeLaps Irsdk_SessionState = iota
	Irsdk_StateRacing     Irsdk_SessionState = iota
	Irsdk_StateCheckered  Irsdk_SessionState = iota
	Irsdk_StateCoolDown   Irsdk_SessionState = iota
)

type Irsdk_CameraState int

const (
	Irsdk_IsSessionScreen Irsdk_CameraState = 0x0001 // the camera tool can only be activated if viewing the session screen (out of car)
	Irsdk_IsScenicActive  Irsdk_CameraState = 0x0002 // the scenic camera is active (no focus car)

	//these can be changed with a broadcast message
	Irsdk_CamToolActive         Irsdk_CameraState = 0x0004
	Irsdk_UIHidden              Irsdk_CameraState = 0x0008
	Irsdk_UseAutoShotSelection  Irsdk_CameraState = 0x0010
	Irsdk_UseTemporaryEdits     Irsdk_CameraState = 0x0020
	Irsdk_UseKeyAcceleration    Irsdk_CameraState = 0x0040
	Irsdk_UseKey10xAcceleration Irsdk_CameraState = 0x0080
	Irsdk_UseMouseAimMode       Irsdk_CameraState = 0x0100
)

//----
//

type Irsdk_varHeader struct {
	Type   irsdk_VarType // irsdk_VarType
	Offset C.int         // offset fron start of buffer row
	Count  C.int         // number of entrys (array)
	// so length in bytes would be irsdk_VarTypeBytes[type] * count

	Pad [1]C.int // (16 byte align)

	Name [IRSDK_MAX_STRING]byte
	Desc [IRSDK_MAX_DESC]byte
	Unit [IRSDK_MAX_STRING]byte // something like "kg/m^2"
}

type irsdk_varBuf struct {
	TickCount C.int    // used to detect changes in data
	BufOffset C.int    // offset from header
	Pad       [2]C.int // (16 byte align)
}

type irsdk_header struct {
	Ver      C.int             // api version 1 for now
	Status   irsdk_StatusField // bitfield using irsdk_StatusField
	TickRate C.int             // ticks per second (60 or 360 etc)

	// session information, updated periodicaly
	SessionInfoUpdate C.int // Incremented when session info changes
	SessionInfoLen    C.int // Length in bytes of session info string
	SessionInfoOffset C.int // Session info, encoded in YAML format

	// State data, output at tickRate
	NumVars         C.int // length of array pointed to by varHeaderOffset
	VarHeaderOffset C.int // offset to Irsdk_varHeader[numVars] array, Describes the variables recieved in varBuf

	NumBuf C.int    // <= IRSDK_MAX_BUFS (3 for now)
	BufLen C.int    // length in bytes for one line
	Pad1   [2]C.int // (16 byte align)
	VarBuf [IRSDK_MAX_BUFS]irsdk_varBuf
}

// sub header used when writing telemetry to disk
type irsdk_diskSubHeader struct {
	sessionStartDate   C.time_t
	sessionStartTime   C.double
	sessionEndTime     C.double
	sessionLapCount    C.int
	sessionRecordCount C.int
}

//----
// Remote controll the sim by sending these windows messages
// camera and replay commands only work when you are out of your car,
// pit commands only work when in your car

type irsdk_BroadcastMsg uint16

const (
	Irsdk_BroadcastCamSwitchPos          irsdk_BroadcastMsg = 0    // car position, group, camera
	Irsdk_BroadcastCamSwitchNum          irsdk_BroadcastMsg = iota // driver #, group, camera
	Irsdk_BroadcastCamSetState           irsdk_BroadcastMsg = iota // irsdk_CameraState, unused, unused
	Irsdk_BroadcastReplaySetPlaySpeed    irsdk_BroadcastMsg = iota // speed, slowMotion, unused
	Irskd_BroadcastReplaySetPlayPosition irsdk_BroadcastMsg = iota // irsdk_RpyPosMode, Frame Number (high, low)
	Irsdk_BroadcastReplaySearch          irsdk_BroadcastMsg = iota // irsdk_RpySrchMode, unused, unused
	Irsdk_BroadcastReplaySetState        irsdk_BroadcastMsg = iota // irsdk_RpyStateMode, unused, unused
	Irsdk_BroadcastReloadTextures        irsdk_BroadcastMsg = iota // irsdk_ReloadTexturesMode, carIdx, unused
	Irsdk_BroadcastChatComand            irsdk_BroadcastMsg = iota // irsdk_ChatCommandMode, subCommand, unused
	Irsdk_BroadcastPitCommand            irsdk_BroadcastMsg = iota // irsdk_PitCommandMode, parameter
	Irsdk_BroadcastTelemCommand          irsdk_BroadcastMsg = iota // irsdk_TelemCommandMode, unused, unused
	Irsdk_BroadcastLast                  irsdk_BroadcastMsg = iota // unused placeholder
)

type irsdk_ChatCommandMode int

const (
	Irsdk_ChatCommand_Macro     irsdk_ChatCommandMode = 0    // pass in a number from 1-15 representing the chat macro to launch
	Irsdk_ChatCommand_BeginChat irsdk_ChatCommandMode = iota // Open up a new chat window
	Irsdk_ChatCommand_Reply     irsdk_ChatCommandMode = iota // Reply to last private chat
	Irsdk_ChatCommand_Cancel    irsdk_ChatCommandMode = iota // Close chat window
)

type irsdk_PitCommandMode int

const (
	Irsdk_PitCommand_Clear      irsdk_PitCommandMode = 0    // Clear all pit checkboxes
	Irsdk_PitCommand_WS         irsdk_PitCommandMode = iota // Clean the winshield, using one tear off
	Irsdk_PitCommand_Fuel       irsdk_PitCommandMode = iota // Add fuel, optionally specify the amount to add in liters or pass '0' to use existing amount
	Irsdk_PitCommand_LF         irsdk_PitCommandMode = iota // Change the left front tire, optionally specifying the pressure in KPa or pass '0' to use existing pressure
	Irsdk_PitCommand_RF         irsdk_PitCommandMode = iota // right front
	Irsdk_PitCommand_LR         irsdk_PitCommandMode = iota // left rear
	Irsdk_PitCommand_RR         irsdk_PitCommandMode = iota // right rear
	Irsdk_PitCommand_ClearTires irsdk_PitCommandMode = iota // Clear tire pit checkboxes
)

type irsdk_TelemCommandMode int

const (
	Irsdk_TelemCommand_Stop    irsdk_TelemCommandMode = 0    // Turn telemetry recording off
	Irsdk_TelemCommand_Start   irsdk_TelemCommandMode = iota // Turn telemetry recording on
	Irsdk_TelemCommand_Restart irsdk_TelemCommandMode = iota // Write current file to disk and start a new one
)

type irsdk_RpyStateMode int

const (
	Irsdk_RpyState_EraseTape irsdk_RpyStateMode = 0    // clear any data in the replay tape
	Irsdk_RpyState_Last      irsdk_RpyStateMode = iota // unused place holder
)

type irsdk_ReloadTexturesMode int

const (
	Irsdk_ReloadTextures_All    irsdk_ReloadTexturesMode = 0    // reload all textuers
	Irsdk_ReloadTextures_CarIdx irsdk_ReloadTexturesMode = iota // reload only textures for the specific carIdx
)

// Search replay tape for events
type irsdk_RpySrchMode int

const (
	Irsdk_RpySrch_ToStart      irsdk_RpySrchMode = 0
	Irsdk_RpySrch_ToEnd        irsdk_RpySrchMode = iota
	Irsdk_RpySrch_PrevSession  irsdk_RpySrchMode = iota
	Irsdk_RpySrch_NextSession  irsdk_RpySrchMode = iota
	Irsdk_RpySrch_PrevLap      irsdk_RpySrchMode = iota
	Irsdk_RpySrch_NextLap      irsdk_RpySrchMode = iota
	Irsdk_RpySrch_PrevFrame    irsdk_RpySrchMode = iota
	Irsdk_RpySrch_NextFrame    irsdk_RpySrchMode = iota
	Irsdk_RpySrch_PrevIncident irsdk_RpySrchMode = iota
	Irsdk_RpySrch_NextIncident irsdk_RpySrchMode = iota
	Irsdk_RpySrch_Last         irsdk_RpySrchMode = iota // unused placeholder
)

type irsdk_RpyPosMode int

const (
	Irsdk_RpyPos_Begin   irsdk_RpyPosMode = 0
	Irsdk_RpyPos_Current irsdk_RpyPosMode = iota
	Irsdk_RpyPos_End     irsdk_RpyPosMode = iota
	Irsdk_RpyPos_Last    irsdk_RpyPosMode = iota // unused placeholder
)

// irsdk_BroadcastCamSwitchPos or irsdk_BroadcastCamSwitchNum camera focus defines
// pass these in for the first parameter to select the 'focus at' types in the camera system.
type irsdk_csMode int

const (
	Irsdk_csFocusAtIncident irsdk_csMode = -3
	Irsdk_csFocusAtLeader   irsdk_csMode = -2
	Irsdk_csFocusAtExiting  irsdk_csMode = -1
	// ctFocusAtDriver + car number...
	Irsdk_csFocusAtDriver irsdk_csMode = 0
)
