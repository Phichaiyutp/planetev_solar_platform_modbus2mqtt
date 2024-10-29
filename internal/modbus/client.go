package modbus

import (
	"time"

	"github.com/grid-x/modbus"
)

type Client struct {
	handler *modbus.RTUOverTCPClientHandler
	client  modbus.Client
}

func NewClient(address string, slaveID byte, timeout time.Duration) (*Client, error) {
	handler := modbus.NewRTUOverTCPClientHandler(address)
	handler.Timeout = timeout
	handler.SlaveID = slaveID

	err := handler.Connect()
	if err != nil {
		return nil, err
	}

	return &Client{
		handler: handler,
		client:  modbus.NewClient(handler),
	}, nil
}

func (c *Client) ReadInputRegisters(address, quantity uint16) ([]byte, error) {
	return c.client.ReadInputRegisters(address, quantity)
}

func (c *Client) Close() error {
	return c.handler.Close()
}
