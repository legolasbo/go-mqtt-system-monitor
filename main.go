package main

import (
	"flag"
	"os"
	"strings"
)

const DEFAULT_CONFIG_PATH = "/etc/msm/config.yml"

func main() {
	list := flag.Bool("list", false, "List all known sensors")
	sensors := flag.String("sensors", "", "Comma separated list of sensors to enable")
	flag.Parse()

	configPath := DEFAULT_CONFIG_PATH
	if len(os.Args) > 1 {
		configPath = os.Args[len(os.Args)-1]
	}

	config := loadConfig(configPath)
	logger := createLogger(config)

	if len(*sensors) > 0 {
		enabledSensors := strings.Split(*sensors, ",")
		config.Sensors = enabledSensors
	}

	if *list {
		listAllSensors(logger)
		return
	}

	daemon := createDaemon(config, logger)

	daemon.Start()
}
