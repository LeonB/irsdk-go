package irsdk

import (
	"bytes"
	"errors"
	"math"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"github.com/leonb/irsdk-go/utils"
)

var (
	ErrEmptySessionData = errors.New("Empty session data")
)

func NewConnection() (*Connection, error) {
	sdk, err := utils.NewIrsdk()
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
	sdk            *utils.Irsdk
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

func (c *Connection) GetHeader() (*utils.Header, error) {
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

func (c *Connection) GetRawSessionData() ([]byte, error) {
	b := c.sdk.GetSessionInfoStr()
	if b == nil {
		return nil, nil
	}

	sep := []byte("\n...")
	pieces := bytes.Split(b, sep)
	if len(pieces) > 0 {
		return pieces[0], nil
	}

	return b, nil
}

func (c *Connection) GetSessionData() (*SessionData, error) {
	b, err := c.GetRawSessionData()
	if err != nil {
		return nil, err
	}

	if b == nil {
		return nil, ErrEmptySessionData
	}

	b = bytesToUtf8(b)
	return NewSessionDataFromBytes(b)
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
