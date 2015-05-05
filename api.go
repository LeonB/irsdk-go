package irsdk

import (
	"bytes"
	"math"
	"time"

	"github.com/leonb/irsdk-go/utils"
)

func NewConnection() (*Connection, error) {
	conn := &Connection{
		timeout: time.Millisecond * time.Duration(math.Ceil(1000.0/60.0)+1.0),
		sdk:     &utils.Irsdk{},
	}

	return conn, nil
}

type Connection struct {
	timeout time.Duration
	sdk     *utils.Irsdk
}

func (c *Connection) Connect() error {
	// If connection was once established: clean it up
	nilTime := time.Time{}
	if c.sdk.GetLastValidTime() != nilTime {
		c.Disconnect()
	}

	err := c.sdk.Startup()
	return err
}

func (c *Connection) IsConnected() bool {
	return c.sdk.IsConnected()
}

func (c *Connection) GetHeader() (*utils.Header, error) {
	return c.sdk.GetHeader()
}

func (c *Connection) GetRawTelemetryData() ([]byte, error) {
	return c.sdk.WaitForDataReady(c.timeout)
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
	data, err := c.sdk.WaitForDataReady(c.timeout)
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
	yamlData, err := c.GetRawSessionData()
	if err != nil {
		return nil, err
	}

	if yamlData != nil {
		return c.BytesToSessionStruct(yamlData)
	}

	return nil, nil
}

func (c *Connection) SendCommand() error {
	return nil
}

func (c *Connection) Disconnect() error {
	c.sdk.Shutdown()
	return nil
}
