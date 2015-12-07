package irsdk

import (
	"bytes"
	"errors"
	"math"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	ErrEmptySessionData = errors.New("Empty session data")
)

func NewConnection() (*Connection, error) {
	sdk, err := NewIrsdk()
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		timeout: time.Millisecond * time.Duration(math.Ceil(1000.0/60.0)+1.0),
		sdk:     sdk,
	}

	return conn, nil
}

type Connection struct {
	timeout        time.Duration
	sdk            *Irsdk
	maxFPS         int
	lastUpdateTime time.Time
}

func (c *Connection) Connect() error {
	// If connection was once established: clean it up
	nilTime := time.Time{}
	if c.sdk.GetLastValidTime() != nilTime {
		c.Disconnect()
	}

	return c.sdk.Startup()
}

func (c *Connection) IsConnected() bool {
	return c.sdk.IsConnected()
}

func (c *Connection) GetHeader() (*Header, error) {
	return c.sdk.GetHeader()
}

func (c *Connection) GetRawTelemetryData() ([]byte, error) {
	return c.WaitForDataReady(c.timeout)
}

func (c *Connection) GetTelemetryData() (*TelemetryData, error) {
	data, err := c.GetRawTelemetryData()
	if err != nil {
		return nil, err
	}

	if data != nil {
		return c.BytesToTelemetryStruct(data)
	}

	return nil, nil
}

func (c *Connection) GetTelemetryDataFiltered(fields []string) (*TelemetryData, error) {
	data, err := c.WaitForDataReady(c.timeout)
	if err != nil {
		return nil, err
	}

	if data != nil {
		return c.BytesToTelemetryStructFiltered(data, fields), nil
	}

	return nil, nil
}

func (c *Connection) GetSessionDataBytes() ([]byte, error) {
	tr := NewTelemetryReader(c.sdk.c.sharedMem)
	return tr.GetSessionDataBytes()
}

func (c *Connection) GetSessionData() (*SessionData, error) {
	return c.sdk.GetSessionData()
}

func (c *Connection) SendCommand() error {
	return nil
}

func (c *Connection) WaitForDataReady(timeOut time.Duration) ([]byte, error) {
	// Check if maxfps specified
	if c.maxFPS == 0 {
		b, err := c.sdk.WaitForDataReady(c.timeout)
		if err == nil {
			c.lastUpdateTime = time.Now()
		}
		return b, err
	}

	// Check if first tick
	nilTime := time.Time{}
	if c.lastUpdateTime == nilTime {
		b, err := c.sdk.WaitForDataReady(c.timeout)
		if err == nil {
			c.lastUpdateTime = time.Now()
		}
		return b, err
	}

	header, err := c.GetHeader()
	if header == nil {
		b, err := c.sdk.WaitForDataReady(c.timeout)
		if err == nil {
			c.lastUpdateTime = time.Now()
		}
		return b, err
	}

	// MaxFPS >= 60, go for max performance
	if c.maxFPS >= int(header.TickRate) {
		b, err := c.sdk.WaitForDataReady(c.timeout)
		if err != nil {
			c.lastUpdateTime = time.Now()
		}
		return b, err
	}

	tickDuration := time.Duration(1*time.Second) / time.Duration(c.maxFPS)
	timeSinceLastUpdate := time.Now().Sub(c.lastUpdateTime)
	timeToWait := tickDuration - timeSinceLastUpdate

	// Sleep to throttle framerate
	time.Sleep(timeToWait)

	// Call non-throttled WaitForDataReady
	b, err := c.sdk.WaitForDataReady(c.timeout)
	if err == nil {
		c.lastUpdateTime = time.Now()
	}
	return b, err
}

func (c *Connection) SetMaxFPS(maxFPS int) {
	c.maxFPS = maxFPS
}

func (c *Connection) Disconnect() error {
	c.sdk.Shutdown()
	return nil
}

// @TODO: should this accept an io.Reader?
func (c *Connection) BytesToTelemetryStruct(data []byte) (*TelemetryData, error) {
	// Create an new struct in the same memory location so reflect values can be
	// cached
	td := NewTelemetryData()
	numVars := c.sdk.GetNumVars()

	for i := 0; i <= numVars; i++ {
		varHeader, err := c.sdk.GetVarHeaderEntry(i)
		if err != nil {
			continue
		}

		if varHeader == nil {
			continue
		}

		td.addVarHeaderData(varHeader, data)
	}

	return td, nil
}

// @TODO: should this accept an io.Reader?
// @TODO: this shouldn't be on the connection because it can also be used by
// disk based telemetry (.ibt)
func (c *Connection) BytesToTelemetryStructFiltered(data []byte, fields []string) *TelemetryData {
	// Create an new struct in the same memory location so reflect values can be
	// cached
	td := NewTelemetryData()
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
			td.addVarHeaderData(varHeader, data)
			continue
		}

		found := false

		for _, v := range fields {
			if v == varHeader.Name {
				// Found varName in fields, skip looping through fields
				found = true
				break
			}
		}

		if found == false {
			// var not in fieds: skip varHeader
			continue
		}

		td.addVarHeaderData(varHeader, data)
	}

	return td
}

// bytesToUtf8 is used to convert stringdata from iRacing to UTF-8 so it can
// safely be used by different encoder methods (json)
func bytesToUtf8(b []byte) []byte {
	isoReader := bytes.NewReader(b)
	// Windows-1252 is a superset of ISO-8859-1
	utf8Reader := transform.NewReader(isoReader, charmap.Windows1252.NewDecoder())

	buf := new(bytes.Buffer)
	buf.ReadFrom(utf8Reader)
	return buf.Bytes()
}
