package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/mem"
)

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
