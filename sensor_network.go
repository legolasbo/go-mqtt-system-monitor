package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/net"
	"log"
	"time"
)

type NetworkTrafficType string

const RX NetworkTrafficType = "RX"
const TX NetworkTrafficType = "TX"

type NetworkSensor struct {
	requestValue chan bool
	value        chan uint64
}

func newNetworkSensor() NetworkSensor {
	return NetworkSensor{
		requestValue: make(chan bool),
		value:        make(chan uint64),
	}
}

func (s *NetworkSensor) getValue() uint64 {
	s.requestValue <- true
	return <-s.value
}

func (s *NetworkSensor) getStringValue() string {
	return fmt.Sprintf("%.3f", float64(s.getValue())/BytesInMegaByte*BitsInByte)
}

func (s *NetworkSensor) start(ntt NetworkTrafficType, updateInterval time.Duration) {
	val, err := net.IOCounters(false)
	if err != nil {
		log.Fatalln(err)
	}

	prev := val[0].BytesRecv
	if ntt == TX {
		prev = val[0].BytesSent
	}
	curr := uint64(0)

	monitor := func(requestValue <-chan bool, value chan<- uint64) {
		t := time.NewTicker(updateInterval)
		for {
			select {
			case <-t.C:
				val, _ = net.IOCounters(false)

				now := val[0].BytesRecv
				if ntt == TX {
					now = val[0].BytesSent
				}

				curr = now - prev
				if int64(now)-int64(prev) < 0 {
					curr = 0
				}
				prev = now
			case <-requestValue:
				value <- curr
			}
		}
	}

	go monitor(s.requestValue, s.value)
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
