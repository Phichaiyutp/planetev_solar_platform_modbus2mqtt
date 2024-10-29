package main

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"modbus-mqtt-service/internal/modbus"
	"modbus-mqtt-service/internal/mqtt"
	"modbus-mqtt-service/internal/processor"
)

func main() {
	// Load configuration from JSON file
	deviceSetting, err := processor.LoadConfig("https://platform.planet-ev.com:8443/modbus/config/51012799/deviceSetting.json")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Modbus client
	address := fmt.Sprintf("%s:%d", deviceSetting.DeviceIp, deviceSetting.DevicePort)
	modbusClient, err := modbus.NewClient(address, byte(deviceSetting.SlaveId), 1*time.Second)
	if err != nil {
		log.Fatalf("Error connecting to Modbus RTU over TCP server: %v", err)
	}
	defer modbusClient.Close()

	// Connect to MQTT broker
	mqttClient := mqtt.NewClient("tcp://api.planetcloud.cloud:1888", "admin", "pca@1234")
	defer mqttClient.Disconnect()
	processorId := fmt.Sprintf("modbus/pnev/%s/%d", deviceSetting.SiteId, deviceSetting.DeviceId)
	// Create processor instance
	proc := processor.NewProcessor(modbusClient, mqttClient, deviceSetting, processorId)

	// Create a new cron instance
	c := cron.New()

	// Schedule a job to run every second
	_, err = c.AddFunc("@every 10s", proc.ProcessData)
	if err != nil {
		log.Fatalf("Error scheduling cron job: %v", err)
	}

	// Start the cron scheduler
	c.Start()

	// Keep the program running indefinitely
	select {}
}
