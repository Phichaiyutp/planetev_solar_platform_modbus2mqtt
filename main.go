package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"modbus-mqtt-service/internal/modbus"
	"modbus-mqtt-service/internal/mqtt"
	"modbus-mqtt-service/internal/processor"

	"github.com/robfig/cron/v3"
)

func main() {
	// Load environment variables
	deviceSettingURL := os.Getenv("DEVICE_SETTING_URL")
	mqttBrokerURL := os.Getenv("MQTT_BROKER_URL")
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")

	// Load configuration from JSON file
	deviceSetting, err := processor.LoadConfig(deviceSettingURL)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
		os.Exit(1) // Ensures non-zero exit on error
	}

	// Initialize Modbus client
	address := fmt.Sprintf("%s:%d", deviceSetting.DeviceIp, deviceSetting.DevicePort)
	modbusClient, err := modbus.NewClient(address, byte(deviceSetting.SlaveId), 1*time.Second)
	if err != nil {
		log.Fatalf("Error connecting to Modbus RTU over TCP server: %v", err)
		os.Exit(1) // Ensures non-zero exit on error
	}
	defer modbusClient.Close()

	// Connect to MQTT broker
	mqttClient := mqtt.NewClient(mqttBrokerURL, mqttUsername, mqttPassword)
	defer mqttClient.Disconnect()
	processorId := fmt.Sprintf("modbus/pnev/%s/%d", deviceSetting.SiteId, deviceSetting.DeviceId)

	// Create processor instance
	proc := processor.NewProcessor(modbusClient, mqttClient, deviceSetting, processorId)

	// Create a new cron instance
	c := cron.New()

	// Schedule a job to run every 10 seconds
	_, err = c.AddFunc("@every 10s", proc.ProcessData)
	if err != nil {
		log.Fatalf("Error scheduling cron job: %v", err)
		os.Exit(1) // Ensures non-zero exit on error
	}

	// Start the cron scheduler
	c.Start()

	// Keep the program running indefinitely
	select {}
}
