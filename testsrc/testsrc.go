package main

import (
	"fmt"
	"github.com/abates/netest"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
	"time"
)

var MTU = 1500
var argParser *flags.Parser
var payload []byte

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalFlag(s string) (err error) {
	d.Duration, err = time.ParseDuration(s)
	return err
}

var options struct {
	BindAddress  string   `short:"a" long:"address" description:"Address to send traffic to" default:"0.0.0.0:0"`
	PollInterval Duration `short:"p" long:"poll" description:"Poll interval to display stats" default:"1s"`
	UseTCP       bool     `short:"t" description:"Use TCP instead of UDP" default:"false"`
}

func getPayload(t netest.PatternType) []byte {
	/* IPv4 Header length is either 20 or 24 bytes depending on whether any
	   options are set.  IPv6 Header length is 40 bytes.  In both cases, the
		 UDP header is 8 bytes */
	if payload == nil {
		length := MTU - 32
		payload = make([]byte, 0)
		for i := 0; i < length; i += len(netest.Patterns[t]) {
			payload = append(payload, netest.Patterns[t]...)
		}
	}
	return payload
}

func main() {
	argParser = flags.NewParser(&options, flags.Default)
	argParser.Usage = "[OPTIONS] <destination address:destination port>"
	args, err := argParser.Parse()

	if err != nil {
		argParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Destination address not specified\n")
		argParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	if options.BindAddress != "0.0.0.0:0" && strings.Count(options.BindAddress, ":") == 0 {
		options.BindAddress = options.BindAddress + ":0"
	}

	var connection *netest.Source
	if options.UseTCP {
		connection, err = netest.NewTCPSource(options.BindAddress, args[0])
	} else {
		connection, err = netest.NewUDPSource(options.BindAddress, args[0])
	}
	defer connection.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create connection %v\n", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(options.PollInterval.Duration)
	var duration time.Duration
	var bytesSent int
	fmt.Printf("\n\n")
	for {
		select {
		case <-ticker.C:
			duration += options.PollInterval.Duration
			fmt.Printf("\033[1A\033[1A")
			fmt.Printf("     TX Rate: %s/s          \n", netest.Humanize(float64(bytesSent)/duration.Seconds()))
			fmt.Printf("    Duration: %-6v Sent: %v          \n", duration, netest.Humanize(float64(bytesSent)))
		default:
			length, err := connection.SendMsg(getPayload(netest.AllOnesPattern))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
				os.Exit(1)
			}
			bytesSent += length
		}
	}
}
