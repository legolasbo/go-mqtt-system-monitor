package main

import (
	"fmt"
	"github.com/hashicorp/mdns"
	"log"
	"strings"
	"time"
)

func hostLookup(host string) {
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %v\n", entry)
		}
	}()

	if strings.HasSuffix(host, ".local") {
		host = host[:len(host)-6]
	}

	// Start the lookup
	params := mdns.DefaultParams(host)
	params.Entries = entriesCh
	params.Timeout = 5 * time.Second
	params.DisableIPv6 = true
	err := mdns.Query(params)
	if err != nil {
		log.Fatalln(err)
		return
	}

	t := time.NewTimer(params.Timeout)
	for {
		select {
		case _ = <-t.C:
			log.Println("closing")
			close(entriesCh)
			return
		case entry := <-entriesCh:
			log.Println(entry)
			break
		}
	}
}
