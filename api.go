package irsdk

import (
	"strings"

	"github.com/leonb/irsdk-go/utils"
	yaml "gopkg.in/yaml.v2"
)

func NewConnection() (*IrConnection, error) {
	conn := &IrConnection{
		timeout: 17, // 1000ms/60f
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

func (c *IrConnection) GetTelemetryData() (*TelemetryData, error) {
	data, err := utils.Irsdk_waitForDataReady(c.timeout)
	if err != nil {
		return nil, err
	}

	if data != nil {
		return BytesToTelemetryData(data), nil
	}

	return nil, nil
}

func (c *IrConnection) GetTelemetryDataFiltered(fields []string) (*TelemetryData, error) {
	data, err := utils.Irsdk_waitForDataReady(c.timeout)
	if err != nil {
		return nil, err
	}

	if data != nil {
		return BytesToTelemetryDataFiltered(data, fields), nil
	}

	return nil, nil
}

func (c *IrConnection) GetSessionData() (string, error) {
	b := utils.Irsdk_getSessionInfoStr()
	if b == nil {
		return "", nil
	}

	s := string(b[:])
	pieces := strings.Split(s, "\n...")
	if len(pieces) > 0 {
		s = pieces[0]
	}
	return s, nil
}

func (c *IrConnection) GetSessionDataStruct() (*SessionData, error) {
	session := SessionData{}
	yamlData, err := c.GetSessionData()
	if err != nil {
		return nil, err
	}

	// Convert yaml to struct
	err = yaml.Unmarshal([]byte(yamlData), &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (c *IrConnection) SendCommand() error {
	return nil
}

func (c *IrConnection) Shutdown() error {
	utils.Irsdk_startup()
	return nil
}
