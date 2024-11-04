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

func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// Load environment variables
	deviceSettingURL := goDotEnvVariable("DEVICE_SETTING_URL")
	mqttBrokerURL := goDotEnvVariable("MQTT_BROKER_URL")
	mqttUsername := goDotEnvVariable("MQTT_USERNAME")
	mqttPassword := goDotEnvVariable("MQTT_PASSWORD")
	mongodbURL := goDotEnvVariable("MONGODB_URL")
	mongodbPort := goDotEnvVariable("MONGODB_PORT")
	mongodbUsername := goDotEnvVariable("MONGODB_USERNAME")
	mongodbPassword := goDotEnvVariable("MONGODB_PASSWORD")
	mongodbDbName := goDotEnvVariable("MONGODB_DB_NAME")
	// Load configuration from JSON file
	deviceSetting, err := processor.LoadConfig(deviceSettingURL)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Modbus client
	if err := modbus.PingPort(deviceSetting.DeviceIp, int(deviceSetting.DevicePort), 3); err != nil {
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
