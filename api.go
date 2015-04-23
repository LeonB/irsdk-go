package irsdk

import (
	"bytes"

	"github.com/leonb/irsdk-go/utils"
)

func NewConnection() (*IrConnection, error) {
	conn := &IrConnection{
		timeout: 18, // (1000ms/60f/s)+1ms(to be sure)
	}

	return conn, conn.Connect()
}

type IrConnection struct {
	timeout int
}

func (c *IrConnection) Connect() error {
	err := utils.Irsdk_startup()
	return err
}

func (c *IrConnection) GetRawTelemetryData() ([]byte, error) {
	return utils.Irsdk_waitForDataReady(c.timeout)
}

func (c *IrConnection) GetTelemetryData() (*TelemetryData, error) {
	data, err := c.GetRawTelemetryData()
	if err != nil {
		return nil, err
	}

	if data != nil {
		return BytesToTelemetryStruct(data), nil
	}

	return nil, nil
}

func (c *IrConnection) GetTelemetryDataFiltered(fields []string) (*TelemetryData, error) {
	data, err := utils.Irsdk_waitForDataReady(c.timeout)
	if err != nil {
		return nil, err
	}

	if data != nil {
		return BytesToTelemetryStructFiltered(data, fields), nil
	}

	return nil, nil
}

func (c *IrConnection) GetRawSessionData() ([]byte, error) {
	b := utils.Irsdk_getSessionInfoStr()
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
		return BytesToSessionStruct(yamlData)
	}

	return nil, nil
}

func (c *IrConnection) SendCommand() error {
	return nil
}

func (c *IrConnection) Shutdown() error {
	utils.Irsdk_startup()
	return nil
}
