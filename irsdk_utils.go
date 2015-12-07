package irsdk

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	MEMMAPFILENAME     = "Local\\IRSDKMemMapFileName"
	BROADCASTMSGNAME   = "IRSDK_BROADCASTMSG"
	DATAVALIDEVENTNAME = "Local\\IRSDKDataValidEvent"
	INT_MAX            = 2147483647
	MEMMAPFILESIZE     = 780 * 1024

	MAX_BUFS   = 4
	MAX_STRING = 32
	// descriptions can be longer than max_string!
	MAX_DESC = 64

	TIMEOUT = time.Duration(time.Second * 30) // timeout after 30 seconds with no communication
)

var (
	ErrInitialize     = errors.New("Failed to initialize")
	ErrDataChanged    = errors.New("Data changed out from under us")
	ErrDisconnected   = errors.New("We probably disconnected")
	ErrNothingChanged = errors.New("Nothing changed this tick")
)

type Irsdk struct {
	isInitialized bool
	lastValidTime time.Time
	lastTickCount int32

	// Syscalls & pointer arithmetic goes into cwrapper
	c *CWrapper
}

func NewIrsdk() (*Irsdk, error) {
	cw, err := NewCWrapper()
	if err != nil {
		return nil, err
	}

	return &Irsdk{c: cw}, nil
}

func (ir *Irsdk) Startup() error {
	var err error

	// Create c wrapper if it doesn't exist
	if ir.c == nil {
		ir.c, err = NewCWrapper()
		if err != nil {
			return err
		}
	}

	// Make sure c wrapper is initialized
	err = ir.c.startup()
	if err != nil {
		ir.c.shutdown()
		return err
	}

	ir.lastTickCount = INT_MAX
	ir.isInitialized = true
	ir.lastValidTime = time.Now()

	return nil
}

func (ir *Irsdk) Shutdown() {
	if ir.c != nil {
		ir.c.shutdown()
		ir.c = nil
	}

	ir.isInitialized = false
	ir.lastTickCount = INT_MAX
}

func (ir *Irsdk) GetNewData() ([]byte, error) {
	var err error

	if !ir.isInitialized {
		err = ir.Startup()
		if err != nil {
			return nil, err
		}
	}

	header := ir.c.header

	// if sim is not active, then no new data
	if (header.Status & StatusConnected) == 0 {
		ir.lastTickCount = INT_MAX
		return nil, nil
	}

	latest := header.GetLatestVarBufN()

	// if newer than last recieved, than report new data
	curTickCount := header.VarBuf[latest].TickCount

	if ir.lastTickCount < curTickCount {
		for count := 0; count < 2; count++ {
			data, err := ir.copyTelemetryData(latest)

			if curTickCount == header.VarBuf[latest].TickCount {
				ir.lastTickCount = curTickCount
				ir.lastValidTime = time.Now()
				return data, err
			}
		}
		// if here, the data changed out from under us.
		return nil, ErrDataChanged
	} else if ir.lastTickCount > header.VarBuf[latest].TickCount {
		// if older than last recieved, than reset, we probably disconnected
		ir.lastTickCount = header.VarBuf[latest].TickCount
		return nil, ErrDisconnected
	}

	// else the same, and nothing changed this tick
	return nil, ErrNothingChanged
}

func (ir *Irsdk) WaitForDataReady(timeOut time.Duration) ([]byte, error) {
	var data []byte
	var err error

	if !ir.isInitialized {
		err = ir.Startup()

		if err != nil {
			// sleep if error
			if timeOut > 0 {
				time.Sleep(timeOut)
			}

			return nil, nil
		}
	}

	// just to be sure, check before we sleep
	data, err = ir.GetNewData()
	if data != nil {
		return data, err
	}

	// sleep till signaled
	ir.c.WaitForDataChange(timeOut)

	// we woke up, so check for data
	data, err = ir.GetNewData()
	return data, err
}

func (ir *Irsdk) IsConnected() bool {
	if ir.isInitialized {
		elapsed := time.Now().Sub(ir.lastValidTime)
		header := ir.c.header
		if (header.Status&StatusConnected) > 0 && (elapsed < TIMEOUT) {
			return true
		}
	}

	return false
}

func (ir *Irsdk) GetVarHeaderEntry(index int) (*VarHeader, error) {
	if ir.isInitialized {
		header := ir.c.header
		if index >= 0 && index < (int)(header.NumVars) {
			return ir.c.getVarHeaderEntry(index)
		}
	}
	return nil, nil
}

// Note: this is a linear search, so cache the results
func (ir *Irsdk) VarNameToIndex(name string) (int, error) {
	if name != "" {
		header := ir.c.header
		numVars := int(header.NumVars)
		for index := 0; index <= numVars; index++ {
			pVar, err := ir.GetVarHeaderEntry(index)
			if err != nil {
				return -1, err
			}
			if pVar != nil && pVar.Name == name {
				return index, nil
			}
		}
	}

	return -1, nil
}

func (ir *Irsdk) VarNameToOffset(name string) (int, error) {
	if name != "" {
		header := ir.c.header
		numVars := int(header.NumVars)
		for index := 0; index <= numVars; index++ {
			pVar, err := ir.GetVarHeaderEntry(index)
			if err != nil {
				return -1, err
			}
			if pVar != nil && pVar.Name == name {
				return int(pVar.Offset), nil
			}
		}
	}

	return -1, nil
}

func (ir *Irsdk) BroadcastMsg(msg BroadcastMsg, var1 uint16, var2 uint16, var3 uint16) error {
	msgID, _ := ir.GetBroadcastMsgID()

	wParam := MAKELONG(uint16(msg), var1)
	lParam := MAKELONG(var2, var3)

	fmt.Println("msgID:", msgID)
	fmt.Println("msg:", msg)
	fmt.Println("var1:", var1)
	fmt.Println("var2:", var2)
	fmt.Println("var3:", var3)
	fmt.Println("wParam", wParam)
	fmt.Println("lParam", lParam)

	wParam2 := uint16(msg) | var1<<16
	lParam2 := var2 | var3<<16

	fmt.Println("wParam2", wParam2)
	fmt.Println("lParam2", lParam2)

	if msgID > 0 && msg >= 0 && msg < BroadcastLast {
		err := ir.c.SendNotifyMessageW(msgID, wParam, lParam)
		err = ir.c.SendNotifyMessageW(msgID, uint32(wParam2), uint32(lParam2))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ir *Irsdk) PadCarNum(num int, zero int) int {
	retVal := num
	numPlace := 1
	if num > 99 {
		numPlace = 3
	} else if num > 9 {
		numPlace = 2
	}
	if zero != 0 {
		numPlace += zero
		retVal = num + 1000*numPlace
	}

	return retVal
}

// Custom functions

func (ir *Irsdk) GetNumVars() int {
	return int(ir.c.header.NumVars)
}

func (ir *Irsdk) GetBroadcastMsgID() (uint, error) {
	return ir.c.RegisterWindowMessageW(BROADCASTMSGNAME)
}

func (ir *Irsdk) copyTelemetryData(varBufN int) ([]byte, error) {
	header := ir.c.header
	bufLen := int(header.BufLen)
	startByte := int(header.VarBuf[varBufN].BufOffset)
	endByte := startByte + bufLen

	fmt.Println("Writing shared mem to memmap.dat")
	ioutil.WriteFile("memmap.dat", ir.c.sharedMem, 0600)

	data := make([]byte, bufLen)
	copy(data, ir.c.sharedMem[startByte:endByte])

	return data, nil
}

func (ir *Irsdk) GetHeader() (*Header, error) {
	return ir.c.getHeader()
}

func (ir *Irsdk) GetSessionData() (*SessionData, error) {
	if !ir.isInitialized {
		err := ir.Startup()
		if err != nil {
			return nil, err
		}
	}

	tr := NewTelemetryReader(ir.c.sharedMem)
	return tr.GetSessionData()
}

func (ir *Irsdk) GetLastValidTime() time.Time {
	return ir.lastValidTime
}

func MAKELONG(lo, hi uint16) uint32 {
	return uint32(uint32(lo) | ((uint32(hi)) << 16))
}

func CToGoString(c []byte) string {
	i := bytes.IndexByte(c, 0)
	return string(c[:i])
}

func CToGoString2(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}
