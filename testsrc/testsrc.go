package main

import (
	"fmt"
	"github.com/abates/netest"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
)

var MTU = 1500
var argParser *flags.Parser
var payload []byte

var options struct {
	BindAddress string `short:"a" long:"address" description:"Address to bind socket to" default:"0.0.0.0:0"`
}

func getPayload(t netest.PatternType) []byte {
	/* IPv4 Header length is either 20 or 24 bytes depending on whether any
	   options are set.  IPv6 Header length is 40 bytes.  In both cases, the
		 UDP header is 8 bytes */
	if payload == nil {
		length := MTU - 32
		payload := make([]byte, length)
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

	connection, err := netest.NewSrc(options.BindAddress, args[0])
	defer connection.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create connection %v\n", err)
		os.Exit(1)
	}

	for {
		err := connection.SendMsg(getPayload(netest.AllOnesPattern))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
			os.Exit(1)
		}
	}
}
