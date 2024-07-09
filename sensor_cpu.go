package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"runtime"
)

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
