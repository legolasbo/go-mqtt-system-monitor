package main

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type HomeAssistantConfig struct {
	Name              string              `json:"name,omitempty"`
	DeviceClass       string              `json:"device_class,omitempty"`
	UnitOfMeasurement string              `json:"unit_of_measurement,omitempty"`
	Device            HomeAssistantDevice `json:"device"`
	ExpireAfter       int                 `json:"expire_after,omitempty"`
	StateTopic        string              `json:"state_topic,omitempty"`
	UniqueId          string              `json:"unique_id,omitempty"`
	ObjectId          string              `json:"object_id,omitempty"`
	StateClass        string              `json:"state_class,omitempty"`
	Icon              string              `json:"icon,omitempty"`
}

type HomeAssistantDevice struct {
	Name        string `json:"name,omitempty"`
	Model       string `json:"model,omitempty"`
	Identifiers string `json:"identifiers,omitempty"`
}

func GetHomeAssistantDevice(conf Config) HomeAssistantDevice {
	caser := cases.Title(language.Dutch)
	return HomeAssistantDevice{
		Name:        caser.String(conf.ClientId),
		Model:       conf.ClientId,
		Identifiers: conf.ClientId,
	}
}

func GetHomeAssistantConfig(daemon Daemon) HomeAssistantConfig {
	return HomeAssistantConfig{
		Name:              daemon.Config.ClientId,
		DeviceClass:       "connectivity",
		UnitOfMeasurement: "None",
		Device:            GetHomeAssistantDevice(daemon.Config),
		ExpireAfter:       DefaultExpireAfter,
		StateTopic:        daemon.stateTopic(),
		UniqueId:          daemon.Config.ClientId,
	}
}
