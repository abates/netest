package main

import (
	"fmt"
	"github.com/abates/netest"
	"os"
	"path"
	"time"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <address:port>\n", path.Base(os.Args[0]))
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	pollInterval := time.Second

	connection, err := netest.NewUdpSink(os.Args[1])
	//defer connection.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create connection: %v\n", err)
		os.Exit(1)
	}

	packetsRead := 0.0
	packetsDropped := 0.0

	count := make(chan uint16)
	sequence := make(chan uint16)

	go func() {
		fmt.Printf("\n\n")
		var lastSequence float64
		var bytesRead uint64
		var duration time.Duration
		ticker := time.NewTicker(pollInterval)
		for {
			select {
			case length := <-count:
				bytesRead += uint64(length)
			case seq := <-sequence:
				s := float64(seq)
				packetsDropped += (s - lastSequence)
				packetsRead++
				lastSequence = s
			case <-ticker.C:
				if packetsRead == 0 {
					packetsRead = 1.0
				}
				duration += pollInterval
				fmt.Printf("\033[1A\033[1A\033[1A")
				fmt.Printf("     RX Rate: %s/s\n", netest.Humanize(float64(bytesRead)/duration))
				fmt.Printf("Success Rate: %.1f\n", (100.0 - (packetsDropped / packetsRead)))
				fmt.Printf("    Duration: %6v Sent: %v\n", duration, netest.Humanize(float64(bytesRead)))
				//bytesRead = 0
			}
		}
	}()

	for {
		packet, err := connection.ReceiveMsg()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed reading from connection %v\n", err)
			os.Exit(1)
		}
		count <- packet.Length
		sequence <- packet.Sequence
	}
}
