package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
)

type BuiltinSensor func() (string, error)
type Sensor struct {
	DeviceClass   string `yaml:"device_class"`
	Description   string `yaml:"description"`
	Id            string `yaml:"id"`
	Name          string `yaml:"name"`
	Script        string `yaml:"script"`
	Builtin       BuiltinSensor
	StartBgSensor func()
	Value         string
	Unit          string `yaml:"unit"`
	StateClass    string `yaml:"state_class"`
	Icon          string `yaml:"icon"`
}

func (s Sensor) HomeAssistantConfig(config Config) (string, HomeAssistantConfig) {
	uniqueId := fmt.Sprintf("%s_%s", config.ClientId, s.Id)
	topic := fmt.Sprintf("homeassistant/sensor/%s/%s/config", config.ClientId, s.Id)
	return topic, HomeAssistantConfig{
		Name:              s.Name,
		DeviceClass:       s.DeviceClass,
		UnitOfMeasurement: s.Unit,
		Device: HomeAssistantDevice{
			Name:        fmt.Sprintf("%s %s", config.ClientId, s.Name),
			Model:       config.ClientId,
			Identifiers: config.ClientId,
		},
		ExpireAfter: DefaultExpireAfter,
		StateTopic:  fmt.Sprintf("%s/%s/%s/%s", config.Prefix, config.ClientId, s.DeviceClass, s.Id),
		UniqueId:    uniqueId,
		ObjectId:    uniqueId,
		StateClass:  s.StateClass,
		Icon:        s.Icon,
	}
}

func (s Sensor) Execute() (Sensor, error) {
	if s.Builtin != nil {
		out, err := s.Builtin()
		if err != nil {
			return s, err
		}
		s.Value = out
		return s, nil
	}

	cmd := exec.Command("bash", "-c", s.Script)
	out, err := cmd.Output()
	if err != nil {
		return s, err
	}

	s.Value = strings.Trim(string(out), " \t\n\r")
	return s, nil
}

func LoadSensors(logger Logger) map[string]Sensor {
	sensors := builtinSensors()

	names := findYamlFiles("default/sensors")
	names = append(names, findYamlFiles(fmt.Sprintf("default/sensors/%s", runtime.GOOS))...)
	names = append(names, findYamlFiles("/etc/msm/sensors")...)

	logger.Debug(fmt.Sprintf("Loading yaml files from:\n %s\n", strings.Join(names, "\n")))

	for _, name := range names {
		b, err := os.ReadFile(name)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		s := Sensor{}
		err = yaml.Unmarshal(b, &s)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		sensors[s.Id] = s
		if s.Icon == "" && strings.HasPrefix(s.Id, "cpu_") {
			s.Icon = cpuIcon()
		}
	}

	return sensors
}

func findYamlFiles(path string) []string {
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	names := make([]string, 0)
	for _, entry := range dir {
		if entry.Type().IsRegular() {
			strings.Split(entry.Name(), ".")
		}
		parts := strings.Split(entry.Name(), ".")
		ext := parts[len(parts)-1]
		if ext == "yml" || ext == "yaml" {
			names = append(names, fmt.Sprintf("%s/%s", path, entry.Name()))
		}
	}
	return names
}

func builtinSensors() map[string]Sensor {
	sensors := make(map[string]Sensor)

	sensors["cpu_cores"] = Sensor{
		Unit:        "",
		Id:          "cpu_cores",
		Name:        "CPU Cores",
		Builtin:     cpuCores,
		Description: "Number of available cpu cores",
		Icon:        "mdi:numeric",
	}
	sensors["cpu_usage"] = Sensor{
		Unit:        "%",
		Id:          "cpu_usage",
		Name:        "CPU Usage",
		Builtin:     cpuPercentage,
		Description: "CPU Usage averaged over all CPU cores in percent",
		StateClass:  "measurement",
		Icon:        cpuIcon(),
	}
	sensors["net_rx_usage"] = Sensor{
		DeviceClass: "data_size",
		Unit:        "GB",
		Id:          "net_rx_usage",
		Name:        "Network RX usage",
		Builtin:     netRxUsage,
		Description: "Total data received over the network in GB",
		Icon:        "mdi:download-network-outline",
	}
	sensors["net_tx_usage"] = Sensor{
		DeviceClass: "data_size",
		Unit:        "GB",
		Id:          "net_tx_usage",
		Name:        "Network TX usage",
		Builtin:     netTxUsage,
		Description: "Total data sent over the network in GB",
		Icon:        "mdi:upload-network-outline",
	}
	rxSens := newNetworkSensor()
	sensors["net_rx"] = Sensor{
		DeviceClass:   "data_rate",
		Unit:          "Mbit/s",
		Id:            "net_rx",
		Name:          "Network RX",
		Builtin:       func() (string, error) { return rxSens.getStringValue(), nil },
		StartBgSensor: func() { rxSens.start(RX, time.Second) },
		Description:   "Data received over the network in Mbit/s",
		Icon:          "mdi:download-network-outline",
	}
	txSens := newNetworkSensor()
	sensors["net_tx"] = Sensor{
		DeviceClass:   "data_rate",
		Unit:          "Mbit/s",
		Id:            "net_tx",
		Name:          "Network TX",
		Builtin:       func() (string, error) { return txSens.getStringValue(), nil },
		StartBgSensor: func() { txSens.start(TX, time.Second) },
		Description:   "Data sent over the network in Mbit/s",
		Icon:          "mdi:upload-network-outline",
	}
	sensors["root_fs_usage"] = Sensor{
		Unit:        "%",
		Id:          "root_fs_usage",
		Name:        "Root FS usage",
		Builtin:     rootFSUsage,
		Description: "Root filesystem usage in percent",
		StateClass:  "measurement",
		Icon:        "mdi:harddisk",
	}
	sensors["available_memory"] = Sensor{
		DeviceClass: "data_size",
		Unit:        "GB",
		Id:          "available_memory",
		Name:        "Available Memory",
		Builtin:     availableMemory,
		Description: "Available memory in GB",
		Icon:        "mdi:memory",
	}
	sensors["occupied_memory"] = Sensor{
		DeviceClass: "data_size",
		Unit:        "GB",
		Id:          "occupied_memory",
		Name:        "Occupied Memory",
		Builtin:     occupiedMemory,
		Description: "Occupied memory in GB",
		Icon:        "mdi:memory",
	}
	sensors["total_memory"] = Sensor{
		DeviceClass: "data_size",
		Unit:        "GB",
		Id:          "total_memory",
		Name:        "Total Memory",
		Builtin:     totalMemory,
		Description: "Total memory in GB",
		Icon:        "mdi:memory",
	}
	sensors["memory_usage"] = Sensor{
		Unit:        "%",
		Id:          "memory_usage",
		Name:        "Memory usage",
		Builtin:     memoryUsage,
		Description: "Memory usage in percent",
		StateClass:  "measurement",
		Icon:        "mdi:memory",
	}

	return sensors
}

func rootFSUsage() (string, error) {
	val, err := disk.Usage("/")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.3f", val.UsedPercent), nil
}

func listAllSensors(logger Logger) {
	fmt.Println("Listing all known sensors. Their names can be used to configure the enabled sensors in the configuration file.")
	sensors := LoadSensors(logger)
	names := make([]string, 0)
	for name := range sensors {
		names = append(names, name)
	}

	sort.Strings(names)

	fmt.Printf("%-20s : %s\n", "Name", "Description")
	fmt.Println("========================================================================")
	for _, name := range names {
		fmt.Printf("%-20s : %s\n", name, sensors[name].Description)
	}
}
