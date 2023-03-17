package main

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Config struct {
	ClientId       string   `yaml:"client-id"`
	HomeAssistant  bool     `yaml:"homeassistant"`
	MQTTBroker     string   `yaml:"mqtt-broker"`
	MQTTBrokerPort int      `yaml:"mqtt_broker-port"`
	MQTTUser       string   `yaml:"mqtt-user"`
	MQTTPassword   string   `yaml:"mqtt-password"`
	Prefix         string   `yaml:"prefix"`
	Sensors        []string `yaml:"sensors"`
	UpdatePeriod   int      `yaml:"update-period"`
	Verbosity      LogLevel `yaml:"log-verbosity"`
}

func loadConfig(path string) Config {
	host, _ := os.Hostname()
	c := Config{
		ClientId:       host,
		MQTTBroker:     "localhost",
		MQTTBrokerPort: 1883,
		Prefix:         "mqtt-system-monitor",
		UpdatePeriod:   10,
		Verbosity:      INFO,
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return c
	}

	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		log.Fatalln(err)
	}

	return c
}
