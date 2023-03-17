package main

import (
	"flag"
	"os"
)

const DEFAULT_CONFIG_PATH = "/etc/msm/config.yml"

func main() {
	list := flag.Bool("list", false, "List all known sensors")
	flag.Parse()

	configPath := DEFAULT_CONFIG_PATH
	if len(os.Args) > 1 {
		configPath = os.Args[len(os.Args)-1]
	}

	config := loadConfig(configPath)
	logger := createLogger(config)

	if *list {
		listAllSensors(logger)
		return
	}

	daemon := createDaemon(config, logger)

	daemon.Start()
}
