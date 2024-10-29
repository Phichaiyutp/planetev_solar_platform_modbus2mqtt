package models

import "time"

// Register defines a structure for a Modbus register with name, address, type, and ratio.
type Register struct {
	Name    string  `json:"name"`
	Address uint16  `json:"address"`
	Type    string  `json:"type"`
	Ratio   float64 `json:"ratio"`
	Unit    string  `json:"unit"`
}

type DeviceSetting struct {
	SiteId      string     `json:"site_id"`
	DeviceIp    string     `json:"device_ip"`
	DevicePort  uint16     `json:"device_port"`
	SlaveId     uint16     `json:"slave_id"`
	DeviceType  uint16     `json:"device_type"`
	DeviceModel string     `json:"device_model"`
	DeviceId    uint16     `json:"device_id"`
	Registers   []Register `json:"registers"`
}

// DataEntry represents each data entry with name, value, and unit.
type DataEntry struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Unit  string      `json:"unit"`
}

// Payload represents the structure of the data to be sent to MQTT.
type Payload struct {
	Timestamp   time.Time   `json:"timestamp"`
	UnixTime    int64       `json:"unix_time"`
	SiteId      string      `json:"site_id"`
	DeviceType  uint16      `json:"device_type"`
	DeviceModel string      `json:"device_model"`
	DeviceId    uint16      `json:"device_id"`
	Data        []DataEntry `json:"data"`
}
