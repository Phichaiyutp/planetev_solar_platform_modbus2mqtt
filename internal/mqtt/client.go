package mqtt

import (
	"encoding/json"
	"log"

	"modbus-mqtt-service/internal/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client mqtt.Client
}

func NewClient(broker, username, password string) *Client {
	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetClientID("modbus-client")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}

	return &Client{client: client}
}

func (c *Client) PublishData(topic string, payload models.Payload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	token := c.client.Publish(topic, 2, true, data)
	token.Wait()
	return token.Error()
}

func (c *Client) Disconnect() {
	c.client.Disconnect(250)
}
