package irsdk

import (
	"bytes"
	"math"
	"time"

	"github.com/leonb/irsdk-go/utils"
)

func NewConnection() (*IrConnection, error) {
	conn := &IrConnection{
		timeout: time.Millisecond * time.Duration(math.Ceil(1000.0/60.0)+1.0),
		sdk:     &utils.Irsdk{},
	}

	return conn, conn.Connect()
}

type IrConnection struct {
	timeout time.Duration
	sdk     *utils.Irsdk
}

func (c *IrConnection) Connect() error {
	err := c.sdk.Startup()
	return err
}

func (c *IrConnection) GetHeader() (*utils.Header, error) {
	return c.sdk.GetHeader()
}

func (c *IrConnection) GetRawTelemetryData() ([]byte, error) {
	return c.sdk.WaitForDataReady(c.timeout)
}

func (c *IrConnection) GetTelemetryData() (*TelemetryData, error) {
	data, err := c.GetRawTelemetryData()
	if err != nil {
		return nil, err
	}

	if data != nil {
		return c.BytesToTelemetryStruct(data)
	}

	return nil, nil
}

func (c *IrConnection) GetTelemetryDataFiltered(fields []string) (*TelemetryData, error) {
	data, err := c.sdk.WaitForDataReady(c.timeout)
	if err != nil {
		return nil, err
	}

	if data != nil {
		return c.BytesToTelemetryStructFiltered(data, fields), nil
	}

	return nil, nil
}

func (c *IrConnection) GetRawSessionData() ([]byte, error) {
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

func (c *IrConnection) GetSessionData() (*SessionData, error) {
	yamlData, err := c.GetRawSessionData()
	if err != nil {
		return nil, err
	}

	if yamlData != nil {
		return c.BytesToSessionStruct(yamlData)
	}

	return nil, nil
}

func (c *IrConnection) SendCommand() error {
	return nil
}

func (c *IrConnection) Shutdown() error {
	c.sdk.Shutdown()
	return nil
}
