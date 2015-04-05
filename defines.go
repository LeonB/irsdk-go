package main

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
	irsdk_stConnected irsdk_StatusField = iota
)

type irsdk_VarType C.int

const (
	// 1 byte
	irsdk_char irsdk_VarType = iota
	irsdk_bool irsdk_VarType = iota

	// 4 bytes
	irsdk_int      irsdk_VarType = iota
	irsdk_bitField irsdk_VarType = iota
	irsdk_float    irsdk_VarType = iota

	// 8 bytes
	irsdk_double irsdk_VarType = iota

	// index, don't use
	irsdk_ETCount = iota
)

// const doesn't work for some reason
var irsdk_VarTypeBytes = [irsdk_ETCount]int{
	1, // irsdk_char
	1, // irsdk_bool

	4, // irsdk_int
	4, // irsdk_bitField
	4, // irsdk_float

	8, // irsdk_double
}

// bit fields
type irsdk_EngineWarnings int

const (
	irsdk_waterTempWarning    irsdk_EngineWarnings = 0x01
	irsdk_fuelPressureWarning irsdk_EngineWarnings = 0x02
	irsdk_oilPressureWarning  irsdk_EngineWarnings = 0x04
	irsdk_engineStalled       irsdk_EngineWarnings = 0x08
	irsdk_pitSpeedLimiter     irsdk_EngineWarnings = 0x10
	irsdk_revLimiterActive    irsdk_EngineWarnings = 0x20
)

type irsdk_Flags uint32

const (
	// global flags
	irsdk_checkered     irsdk_Flags = 0x00000001
	irsdk_white         irsdk_Flags = 0x00000002
	irsdk_green         irsdk_Flags = 0x00000004
	irsdk_yellow        irsdk_Flags = 0x00000008
	irsdk_red           irsdk_Flags = 0x00000010
	irsdk_blue          irsdk_Flags = 0x00000020
	irsdk_debris        irsdk_Flags = 0x00000040
	irsdk_crossed       irsdk_Flags = 0x00000080
	irsdk_yellowWaving  irsdk_Flags = 0x00000100
	irsdk_oneLapToGreen irsdk_Flags = 0x00000200
	irsdk_greenHeld     irsdk_Flags = 0x00000400
	irsdk_tenToGo       irsdk_Flags = 0x00000800
	irsdk_fiveToGo      irsdk_Flags = 0x00001000
	irsdk_randomWaving  irsdk_Flags = 0x00002000
	irsdk_caution       irsdk_Flags = 0x00004000
	irsdk_cautionWaving irsdk_Flags = 0x00008000

	// drivers black flags
	irsdk_black      irsdk_Flags = 0x00010000
	irsdk_disqualify irsdk_Flags = 0x00020000
	irsdk_servicible irsdk_Flags = 0x00040000 // car is allowed service (not a flag)
	irsdk_furled     irsdk_Flags = 0x00080000
	irsdk_repair     irsdk_Flags = 0x00100000

	// start lights
	irsdk_startHidden irsdk_Flags = 0x10000000
	irsdk_startReady  irsdk_Flags = 0x20000000
	irsdk_startSet    irsdk_Flags = 0x40000000
	irsdk_startGo     irsdk_Flags = 0x80000000
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

type irsdk_SessionState int

const (
	irsdk_StateInvalid    irsdk_SessionState = iota
	irsdk_StateGetInCar   irsdk_SessionState = iota
	irsdk_StateWarmup     irsdk_SessionState = iota
	irsdk_StateParadeLaps irsdk_SessionState = iota
	irsdk_StateRacing     irsdk_SessionState = iota
	irsdk_StateCheckered  irsdk_SessionState = iota
	irsdk_StateCoolDown   irsdk_SessionState = iota
)

type irsdk_CameraState int

const (
	irsdk_IsSessionScreen irsdk_CameraState = 0x0001 // the camera tool can only be activated if viewing the session screen (out of car)
	irsdk_IsScenicActive  irsdk_CameraState = 0x0002 // the scenic camera is active (no focus car)

	//these can be changed with a broadcast message
	irsdk_CamToolActive         irsdk_CameraState = 0x0004
	irsdk_UIHidden              irsdk_CameraState = 0x0008
	irsdk_UseAutoShotSelection  irsdk_CameraState = 0x0010
	irsdk_UseTemporaryEdits     irsdk_CameraState = 0x0020
	irsdk_UseKeyAcceleration    irsdk_CameraState = 0x0040
	irsdk_UseKey10xAcceleration irsdk_CameraState = 0x0080
	irsdk_UseMouseAimMode       irsdk_CameraState = 0x0100
)

//----
//

type irsdk_varHeader struct {
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
	VarHeaderOffset C.int // offset to irsdk_varHeader[numVars] array, Describes the variables recieved in varBuf

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
	irsdk_BroadcastCamSwitchPos          irsdk_BroadcastMsg = 0    // car position, group, camera
	irsdk_BroadcastCamSwitchNum          irsdk_BroadcastMsg = iota // driver #, group, camera
	irsdk_BroadcastCamSetState           irsdk_BroadcastMsg = iota // irsdk_CameraState, unused, unused
	irsdk_BroadcastReplaySetPlaySpeed    irsdk_BroadcastMsg = iota // speed, slowMotion, unused
	irskd_BroadcastReplaySetPlayPosition irsdk_BroadcastMsg = iota // irsdk_RpyPosMode, Frame Number (high, low)
	irsdk_BroadcastReplaySearch          irsdk_BroadcastMsg = iota // irsdk_RpySrchMode, unused, unused
	irsdk_BroadcastReplaySetState        irsdk_BroadcastMsg = iota // irsdk_RpyStateMode, unused, unused
	irsdk_BroadcastReloadTextures        irsdk_BroadcastMsg = iota // irsdk_ReloadTexturesMode, carIdx, unused
	irsdk_BroadcastChatComand            irsdk_BroadcastMsg = iota // irsdk_ChatCommandMode, subCommand, unused
	irsdk_BroadcastPitCommand            irsdk_BroadcastMsg = iota // irsdk_PitCommandMode, parameter
	irsdk_BroadcastTelemCommand          irsdk_BroadcastMsg = iota // irsdk_TelemCommandMode, unused, unused
	irsdk_BroadcastLast                  irsdk_BroadcastMsg = iota // unused placeholder
)

type irsdk_ChatCommandMode int

const (
	irsdk_ChatCommand_Macro     irsdk_ChatCommandMode = 0    // pass in a number from 1-15 representing the chat macro to launch
	irsdk_ChatCommand_BeginChat irsdk_ChatCommandMode = iota // Open up a new chat window
	irsdk_ChatCommand_Reply     irsdk_ChatCommandMode = iota // Reply to last private chat
	irsdk_ChatCommand_Cancel    irsdk_ChatCommandMode = iota // Close chat window
)

type irsdk_PitCommandMode int

const (
	irsdk_PitCommand_Clear      irsdk_PitCommandMode = 0    // Clear all pit checkboxes
	irsdk_PitCommand_WS         irsdk_PitCommandMode = iota // Clean the winshield, using one tear off
	irsdk_PitCommand_Fuel       irsdk_PitCommandMode = iota // Add fuel, optionally specify the amount to add in liters or pass '0' to use existing amount
	irsdk_PitCommand_LF         irsdk_PitCommandMode = iota // Change the left front tire, optionally specifying the pressure in KPa or pass '0' to use existing pressure
	irsdk_PitCommand_RF         irsdk_PitCommandMode = iota // right front
	irsdk_PitCommand_LR         irsdk_PitCommandMode = iota // left rear
	irsdk_PitCommand_RR         irsdk_PitCommandMode = iota // right rear
	irsdk_PitCommand_ClearTires irsdk_PitCommandMode = iota // Clear tire pit checkboxes
)

type irsdk_TelemCommandMode int

const (
	irsdk_TelemCommand_Stop    irsdk_TelemCommandMode = 0    // Turn telemetry recording off
	irsdk_TelemCommand_Start   irsdk_TelemCommandMode = iota // Turn telemetry recording on
	irsdk_TelemCommand_Restart irsdk_TelemCommandMode = iota // Write current file to disk and start a new one
)

type irsdk_RpyStateMode int

const (
	irsdk_RpyState_EraseTape irsdk_RpyStateMode = 0    // clear any data in the replay tape
	irsdk_RpyState_Last      irsdk_RpyStateMode = iota // unused place holder
)

type irsdk_ReloadTexturesMode int

const (
	irsdk_ReloadTextures_All    irsdk_ReloadTexturesMode = 0    // reload all textuers
	irsdk_ReloadTextures_CarIdx irsdk_ReloadTexturesMode = iota // reload only textures for the specific carIdx
)

// Search replay tape for events
type irsdk_RpySrchMode int

const (
	irsdk_RpySrch_ToStart      irsdk_RpySrchMode = 0
	irsdk_RpySrch_ToEnd        irsdk_RpySrchMode = iota
	irsdk_RpySrch_PrevSession  irsdk_RpySrchMode = iota
	irsdk_RpySrch_NextSession  irsdk_RpySrchMode = iota
	irsdk_RpySrch_PrevLap      irsdk_RpySrchMode = iota
	irsdk_RpySrch_NextLap      irsdk_RpySrchMode = iota
	irsdk_RpySrch_PrevFrame    irsdk_RpySrchMode = iota
	irsdk_RpySrch_NextFrame    irsdk_RpySrchMode = iota
	irsdk_RpySrch_PrevIncident irsdk_RpySrchMode = iota
	irsdk_RpySrch_NextIncident irsdk_RpySrchMode = iota
	irsdk_RpySrch_Last         irsdk_RpySrchMode = iota // unused placeholder
)

type irsdk_RpyPosMode int

const (
	irsdk_RpyPos_Begin   irsdk_RpyPosMode = 0
	irsdk_RpyPos_Current irsdk_RpyPosMode = iota
	irsdk_RpyPos_End     irsdk_RpyPosMode = iota
	irsdk_RpyPos_Last    irsdk_RpyPosMode = iota // unused placeholder
)

// irsdk_BroadcastCamSwitchPos or irsdk_BroadcastCamSwitchNum camera focus defines
// pass these in for the first parameter to select the 'focus at' types in the camera system.
type irsdk_csMode int

const (
	irsdk_csFocusAtIncident irsdk_csMode = -3
	irsdk_csFocusAtLeader   irsdk_csMode = -2
	irsdk_csFocusAtExiting  irsdk_csMode = -1
	// ctFocusAtDriver + car number...
	irsdk_csFocusAtDriver irsdk_csMode = 0
)
