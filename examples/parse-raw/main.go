package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	packet "github.com/konradit/crsf/pkg/crsfpacket"
)

func main() {
	inputFile := flag.String("i", "data.bin", "input file")
	flag.Parse()

	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist\n", *inputFile)
		os.Exit(1)
	}

	input, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		os.Exit(1)
	}
	defer input.Close()

	buf := make([]byte, 100)
	n, err := input.Read(buf)
	if err != nil {
		fmt.Printf("Failed to read from file: %v\n", err)
		os.Exit(1)
	}

	buf = buf[:n]
	if buf[0] == packet.SyncByte { // sync byte.
		if len(buf) < 22 {
			fmt.Println("Not enough bytes to parse packet")
			os.Exit(1)
		}
		// validate the size
		packetSize := buf[1] - 2
		packetType := packet.PacketType(buf[2])

		expectedSize := byte(0x16)                 // rc channels size.
		expectedType := packet.FrameChannelsPacked // rc channels type.

		if packetSize != expectedSize || packetType != expectedType {
			fmt.Println("Unexpected type or size.")
			os.Exit(1)
		}

		stripped := buf[:packetSize+4]

		content, err := packet.ParseFrame(stripped)
		if err != nil {
			fmt.Printf("Failed to parse frame: %v\n", err)
			os.Exit(1)
		}

		asJSON, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal to JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Content: %s\n", asJSON)

		channels := packet.UnpackChannels(content.Payload)
		fmt.Printf("Channels:\n%v\n", channels)
	}
}
