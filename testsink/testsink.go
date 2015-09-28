package main

import (
	"fmt"
	"github.com/abates/netest"
	"github.com/jessevdk/go-flags"
	"os"
	"time"
)

var argParser *flags.Parser

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalFlag(s string) (err error) {
	d.Duration, err = time.ParseDuration(s)
	return err
}

var options struct {
	PollInterval Duration `short:"p" long:"poll" description:"Poll interval to display stats" default:"1s"`
	UseTCP       bool     `short:"t" description:"Use TCP instead of UDP" default:"false"`
}

func main() {
	argParser = flags.NewParser(&options, flags.Default)
	argParser.Usage = "[OPTIONS] <address:port>"
	args, err := argParser.Parse()

	if err != nil {
		argParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Address not specified\n")
		argParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	var connection *netest.Sink
	if options.UseTCP {
		connection, err = netest.NewTCPSink(args[0])
	} else {
		connection, err = netest.NewUDPSink(args[0])
	}
	defer connection.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create connection: %v\n", err)
		argParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	packetsRead := 0.0
	packetsDropped := 0.0

	count := make(chan uint16)
	sequence := make(chan uint16)

	go func() {
		fmt.Printf("\n\n\n")
		var lastSequence float64
		var bytesRead uint64
		var duration time.Duration
		ticker := time.NewTicker(options.PollInterval.Duration)
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
				duration += options.PollInterval.Duration
				fmt.Printf("\033[1A\033[1A\033[1A")
				fmt.Printf("     RX Rate: %v/s      \n", netest.Humanize(float64(bytesRead)/duration.Seconds()))
				fmt.Printf("Success Rate: %-.1f\n", (100.0 - (packetsDropped / packetsRead)))
				fmt.Printf("    Duration: %-6v Received: %-10v\n", duration, netest.Humanize(float64(bytesRead)))
			}
		}
	}()

	for {
		packet := connection.ReceiveMsg()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed reading from connection %v\n", err)
			os.Exit(1)
		}
		count <- packet.Length
		sequence <- packet.Sequence
	}
}
