package main

const DefaultExpireAfter = 60

type HomeAssistantConfig struct {
	Name              string              `json:"name,omitempty"`
	Class             string              `json:"class,omitempty"`
	UnitOfMeasurement string              `json:"unit_of_measurement,omitempty"`
	Device            HomeAssistantDevice `json:"device"`
	ExpireAfter       int                 `json:"expire_after,omitempty"`
	StateTopic        string              `json:"state_topic,omitempty"`
	UniqueId          string              `json:"unique_id,omitempty"`
	ObjectId          string              `json:"object_id,omitempty"`
}

type HomeAssistantDevice struct {
	Name        string `json:"name,omitempty"`
	Model       string `json:"model,omitempty"`
	Identifiers string `json:"identifiers,omitempty"`
}

func GetHomeAssistantConfig(daemon Daemon) HomeAssistantConfig {
	return HomeAssistantConfig{
		Name:              daemon.Config.ClientId,
		Class:             "connectivity",
		UnitOfMeasurement: "None",
		Device: HomeAssistantDevice{
			Name:        daemon.Config.ClientId,
			Model:       daemon.Config.ClientId,
			Identifiers: daemon.Config.ClientId,
		},
		ExpireAfter: DefaultExpireAfter,
		StateTopic:  daemon.stateTopic(),
		UniqueId:    daemon.Config.ClientId,
	}
}
