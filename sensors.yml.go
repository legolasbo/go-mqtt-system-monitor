package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
)

const BitsInByte = 8
const BytesInKiloByte = 1000
const BytesInMegaByte = 1000 * BytesInKiloByte
const BytesInGigaByte = 100 * BytesInMegaByte

type BuiltinSensor func() (string, error)
type Sensor struct {
	DeviceClass string `yaml:"device_class"`
	Description string `yaml:"description"`
	Id          string `yaml:"id"`
	Name        string `yaml:"name"`
	Script      string `yaml:"script"`
	Builtin     BuiltinSensor
	Value       string
	Unit        string `yaml:"unit"`
	StateClass  string `yaml:"state_class"`
	Icon        string `yaml:"icon"`
}

func (s Sensor) HomeAssistantConfig(config Config) (string, HomeAssistantConfig) {
	uniqueId := fmt.Sprintf("%s_%s", config.ClientId, s.Id)
	topic := fmt.Sprintf("homeassistant/sensor/%s/config", uniqueId)
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
	sensors["net_rx"] = Sensor{
		DeviceClass: "data_rate",
		Unit:        "Mbit/s",
		Id:          "net_rx",
		Name:        "Network RX",
		Builtin: func() (string, error) {
			return fmt.Sprintf("%.3f", float64(netRxBytesLastSec)/BytesInMegaByte*BitsInByte), nil
		},
		Description: "Data received over the network in Mbit/s",
		Icon:        "mdi:download-network-outline",
	}
	sensors["net_tx"] = Sensor{
		DeviceClass: "data_rate",
		Unit:        "Mbit/s",
		Id:          "net_tx",
		Name:        "Network TX",
		Builtin: func() (string, error) {
			return fmt.Sprintf("%.3f", float64(netTxBytesLastSec)/BytesInMegaByte*BitsInByte), nil
		},
		Description: "Data sent over the network in Mbit/s",
		Icon:        "mdi:upload-network-outline",
	}
	sensors["root_fs_usage"] = Sensor{
		Unit:        "%",
		Id:          "root_fs_usage",
		Name:        "Root FS usage",
		Builtin:     rootFSUsage,
		Description: "Root filesystem usage in percent",
		StateClass:  "measurement",
		Icon:        "mdi:hardisk",
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

func cpuIcon() string {
	return fmt.Sprintf("mdi:cpu-%d-bit", wordSize())
}

func wordSize() int {
	b64Archs := []string{"amd64", "arm64", "arm64be", "loong64", "mips64", "mips64le", "ppc64", "ppc64le", "riscv64", "s390x", "sparc64", "wasm"}
	for _, arch := range b64Archs {
		if runtime.GOARCH == arch {
			return 64
		}
	}
	return 32
}

func rootFSUsage() (string, error) {
	val, err := disk.Usage("/")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.3f", val.UsedPercent), nil
}

var netRxBytesLastSec uint64 = 0
var netTxBytesLastSec uint64 = 0

func senseNetworkSpeed(d *Daemon) {
	d.Logger.Info("Starting network speed background sensor")
	val, err := net.IOCounters(false)
	if err != nil {
		d.Logger.Error("Failed to start network speed background sensor!")
		d.Logger.Error(err.Error())
		return
	}
	tx := val[0].BytesSent
	rx := val[0].BytesRecv

	t := time.NewTicker(time.Second)
	for range t.C {
		val, _ = net.IOCounters(false)
		netTxBytesLastSec = val[0].BytesSent - tx
		netRxBytesLastSec = val[0].BytesRecv - rx

		tx = val[0].BytesSent
		rx = val[0].BytesRecv
	}
}

func netRxUsage() (string, error) {
	val, err := net.IOCounters(false)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.3f", float64(val[0].BytesRecv)/BytesInGigaByte), nil
}

func netTxUsage() (string, error) {
	val, err := net.IOCounters(false)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.3f", float64(val[0].BytesSent)/BytesInGigaByte), nil
}

func cpuCores() (string, error) {
	val, err := cpu.Info()
	if err != nil {
		return "", err
	}
	cores := 0
	for _, stat := range val {
		cores += int(stat.Cores)
	}
	return fmt.Sprintf("%d", cores), nil
}

func cpuPercentage() (string, error) {
	val, err := cpu.Percent(0, false)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%.3f", val[0]), nil
}

func availableMemory() (string, error) {
	val, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.1f", float64(val.Available)/BytesInGigaByte), nil
}

func occupiedMemory() (string, error) {
	val, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.1f", float64(val.Used)/BytesInGigaByte), nil
}

func memoryUsage() (string, error) {
	val, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.3f", val.UsedPercent), nil
}

func totalMemory() (string, error) {
	val, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.1f", float64(val.Total)/BytesInGigaByte), nil
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
