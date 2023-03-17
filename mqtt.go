package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func getMQTTClient(config Config, logger Logger) (mqtt.Client, error) {

	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		logger.Info("Connected")
	}

	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		logger.Warn(fmt.Sprintf("Connect lost: %v", err))
		logger.Info("Reconnecting...")
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Fatalln(token.Error())
		}
	}

	logger.Info("Connecting...")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.MQTTBroker, config.MQTTBrokerPort))
	opts.SetClientID(config.ClientId)
	opts.SetUsername(config.MQTTUser)
	opts.SetPassword(config.MQTTPassword)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}
