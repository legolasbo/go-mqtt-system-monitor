package main

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Daemon struct {
	Config  Config
	Logger  Logger
	client  mqtt.Client
	sensors map[string]Sensor
}

func (d *Daemon) Start() {
	d.Logger.Info("Starting daemon...")

	var err error
	d.client, err = getMQTTClient(d.Config, d.Logger)
	if err != nil {
		log.Fatalln(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	expireTicker := time.NewTicker((DefaultExpireAfter - 1) * time.Second)
	updateTicker := time.NewTicker(time.Duration(d.Config.UpdatePeriod) * time.Second)

	d.configureHomeAssistant()
	for {
		select {
		case <-sigs:
			d.publishState(false)
			return
		case <-expireTicker.C:
			go d.configureHomeAssistant()
			break
		case <-updateTicker.C:
			go d.publishState(true)
			go d.readSensors()
			break
		}
	}
}

func (d *Daemon) configureHomeAssistant() {
	topic := fmt.Sprintf("homeassistant/binary_sensor/%s/config", d.Config.ClientId)
	d.publishHomeAssistantConfig(topic, GetHomeAssistantConfig(*d))

	for _, sensor := range d.sensors {
		d.publishHomeAssistantConfig(sensor.HomeAssistantConfig(d.Config))
	}
}

func (d *Daemon) publishHomeAssistantConfig(topic string, config HomeAssistantConfig) {
	val, err := json.Marshal(config)
	if err != nil {
		d.Logger.Warn(err.Error())
		return
	}
	d.publish(topic, string(val))
}

func (d *Daemon) readSensors() {
	for _, sensor := range d.sensors {
		value, err := sensor.Execute()
		if err != nil {
			d.Logger.Error(err.Error())
			continue
		}
		d.publish(fmt.Sprintf("%s/%s/%s/%s", d.Config.Prefix, d.Config.ClientId, value.Class, value.Id), value.Value)
	}
}

func (d *Daemon) publishState(state bool) bool {
	if state {
		d.publishAndWait(d.stateTopic(), "ON")
		return true
	}

	d.publishAndWait(d.stateTopic(), "OFF")
	return true
}

func (d *Daemon) stateTopic() string {
	return fmt.Sprintf("%s/%s/state", d.Config.Prefix, d.Config.ClientId)
}

func (d *Daemon) publish(topic string, msg string) mqtt.Token {
	d.Logger.Debug(fmt.Sprintf("Publishing to %s: %s", topic, msg))
	return d.client.Publish(topic, 1, false, msg)
}

func (d *Daemon) publishAndWait(topic string, msg string) {
	token := d.publish(topic, msg)
	token.Wait()
}

func (d *Daemon) filterSensors() {
	if len(d.Config.Sensors) == 0 {
		return
	}

	filtered := make(map[string]Sensor)
	for _, name := range d.Config.Sensors {
		s, ok := d.sensors[name]
		if !ok {
			d.Logger.Warn(fmt.Sprintf("Unknown name: %s", name))
			continue
		}

		filtered[name] = s
	}
	d.sensors = filtered
}

func createDaemon(config Config, logger Logger) Daemon {
	_, err := net.LookupIP(config.MQTTBroker)
	if err != nil {
		hostLookup(config.MQTTBroker)
	}

	d := Daemon{Config: config, Logger: logger}

	d.sensors = LoadSensors(logger)
	d.filterSensors()

	return d
}
