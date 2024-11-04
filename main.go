package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"modbus-mqtt-service/internal/modbus"
	"modbus-mqtt-service/internal/mongodb"
	"modbus-mqtt-service/internal/mqtt"
	"modbus-mqtt-service/internal/processor"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func loadEnvVariables() (string, string, string, string, string, string, string, string, string) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	deviceSettingURL := os.Getenv("DEVICE_SETTING_URL")
	mqttBrokerURL := os.Getenv("MQTT_BROKER_URL")
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	mongodbURL := os.Getenv("MONGODB_URL")
	mongodbPort := os.Getenv("MONGODB_PORT")
	mongodbUsername := os.Getenv("MONGODB_USERNAME")
	mongodbPassword := os.Getenv("MONGODB_PASSWORD")
	mongodbDbName := os.Getenv("MONGODB_DB_NAME")

	if deviceSettingURL == "" {
		log.Fatalf("DEVICE_SETTING_URL is not set in environment variables")
	}

	return deviceSettingURL, mqttBrokerURL, mqttUsername, mqttPassword, mongodbURL, mongodbPort, mongodbUsername, mongodbPassword, mongodbDbName
}

func main() {
	// Load environment variables
	deviceSettingURL, mqttBrokerURL, mqttUsername, mqttPassword, mongodbURL, mongodbPort, mongodbUsername, mongodbPassword, mongodbDbName := loadEnvVariables()

	// Load configuration from JSON file
	deviceSetting, err := processor.LoadConfig(deviceSettingURL)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Modbus client
	if err := processor.PingPort(deviceSetting.DeviceIp, int(deviceSetting.DevicePort), 3); err != nil {
		log.Fatalf("Error connecting to Modbus Converter: %v", err)
	}

	address := fmt.Sprintf("%s:%d", deviceSetting.DeviceIp, deviceSetting.DevicePort)
	modbusClient, err := modbus.NewClient(address, byte(deviceSetting.SlaveId), 1*time.Second)
	if err != nil {
		log.Fatalf("Error connecting to Modbus RTU over TCP server: %v", err)
	}
	defer modbusClient.Close()

	// Connect to MQTT broker
	mqttClient := mqtt.NewClient(mqttBrokerURL, mqttUsername, mqttPassword)
	defer mqttClient.Disconnect()

	// MongoDB connection
	mongodbClient, err := mongodb.GetMongoClient(mongodbUsername, mongodbPassword, mongodbURL, mongodbPort, mongodbDbName)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer mongodbClient.CloseClient()

	processorId := fmt.Sprintf("modbus/pnev/%s/%d", deviceSetting.SiteId, deviceSetting.DeviceId)

	// Create processor instance
	proc := processor.NewProcessor(modbusClient, mqttClient, mongodbClient, deviceSetting, processorId)

	// Create a new cron instance
	c := cron.New()

	// Schedule a job to run every 10 seconds
	_, err = c.AddFunc("@every 10s", proc.ProcessData)
	if err != nil {
		log.Fatalf("Error scheduling cron job: %v", err)
	}

	// Start the cron scheduler
	c.Start()

	// Keep the program running indefinitely
	defer c.Stop() // Ensure cron stops when exiting
	select {}
}
