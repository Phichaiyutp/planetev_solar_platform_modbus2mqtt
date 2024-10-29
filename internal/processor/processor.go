package processor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"modbus-mqtt-service/internal/modbus"
	"modbus-mqtt-service/internal/models"
	"modbus-mqtt-service/internal/mqtt"
)

type Processor struct {
	modbusClient  *modbus.Client
	mqttClient    *mqtt.Client
	deviceSetting models.DeviceSetting
	processorId   string
}

func NewProcessor(modbusClient *modbus.Client, mqttClient *mqtt.Client, deviceSetting models.DeviceSetting, processorId string) *Processor {
	// Sort registers by address
	sort.Slice(deviceSetting.Registers, func(i, j int) bool {
		return deviceSetting.Registers[i].Address < deviceSetting.Registers[j].Address
	})

	return &Processor{
		modbusClient:  modbusClient,
		mqttClient:    mqttClient,
		deviceSetting: deviceSetting,
		processorId:   processorId,
	}
}

// LoadConfig fetches the device settings from the provided URL and unmarshals it into a DeviceSetting struct.
func LoadConfig(url string) (models.DeviceSetting, error) {
	// Send a GET request to the URL
	response, err := http.Get(url)
	if err != nil {
		return models.DeviceSetting{}, fmt.Errorf("failed to fetch data: %w", err)
	}
	// Ensure the response body is closed after function execution
	defer response.Body.Close()

	// Check if the HTTP status code is OK
	if response.StatusCode != http.StatusOK {
		return models.DeviceSetting{}, fmt.Errorf("failed to fetch data: %s", response.Status)
	}

	// Decode the JSON response into the deviceSetting struct
	var deviceSetting models.DeviceSetting
	err = json.NewDecoder(response.Body).Decode(&deviceSetting)
	if err != nil {
		return models.DeviceSetting{}, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return deviceSetting, nil
}

// groupRegisters groups registers with addresses that are close to each other (difference <= 2).
func groupRegisters(registers []models.Register) [][]models.Register {
	if len(registers) == 0 {
		return nil
	}

	var groups [][]models.Register
	currentGroup := []models.Register{registers[0]}

	for i := 1; i < len(registers); i++ {
		// If the difference between current and previous register address is <= 2,
		// add to current group
		if registers[i].Address-registers[i-1].Address <= 2 {
			currentGroup = append(currentGroup, registers[i])
		} else {
			// Start a new group
			groups = append(groups, currentGroup)
			currentGroup = []models.Register{registers[i]}
		}
	}

	// Add the last group if it exists
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

func (p *Processor) ProcessData() {
	groupedRegisters := groupRegisters(p.deviceSetting.Registers)
	payload := models.Payload{
		Timestamp:   time.Now(),
		UnixTime:    time.Now().Unix(),
		SiteId:      p.deviceSetting.SiteId,
		DeviceType:  p.deviceSetting.DeviceType,
		DeviceModel: p.deviceSetting.DeviceModel,
		DeviceId:    p.deviceSetting.DeviceId,
		Data:        []models.DataEntry{},
	}

	for _, group := range groupedRegisters {
		startAddr := group[0].Address
		endAddr := group[len(group)-1].Address

		results, err := p.modbusClient.ReadInputRegisters(startAddr, endAddr-startAddr+2)
		if err != nil {
			log.Fatalf("Error reading from modbus device: %v", err)
		}

		processRegisterData(group, results, &payload)
	}
	topic := p.processorId
	if err := p.mqttClient.PublishData(topic, payload); err != nil {
		log.Fatalf("Error sending to MQTT server: %v", err)
	}
}
