package modbus

import (
	"fmt"
	"net"
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

func PingPort(ip string, port int, timeout time.Duration) error {
	address := fmt.Sprintf("%s:%d", ip, port)

	// Try to establish a udp connection to the specified address with a timeout
	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", address, err)
	}
	defer conn.Close() // Ensure the connection is closed after function execution

	fmt.Printf("Successfully connected to %s\n", address)
	return nil
}
