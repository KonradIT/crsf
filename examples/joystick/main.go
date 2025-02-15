package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artman41/vjoy"
	"github.com/konradit/crsf"
)

func getJoystick() *vjoy.Device {
	var currentID uint = 1
	var maxJoyID uint = 30

	dev, err := vjoy.Acquire(currentID)
	if err != nil {
		currentID++
	}
	for err == vjoy.ErrDeviceAlreadyOwned && currentID <= maxJoyID {
		dev, err = vjoy.Acquire(currentID)
		currentID++
	}
	if err != nil {
		fmt.Println("Failed to acquire joystick")
		os.Exit(1)
	}
	return dev
}

func scaleToJoystick(value uint16, maxValue uint16) int {
	// Normalize to -1 to 1
	normalized := (float64(value)/float64(maxValue))*2 - 1
	// Scale to -0x4000..0x3fff (-16384 to 16383)
	// This is how vJoy interprets the values.
	return int(normalized * 16383)
}

func main() {
	fmt.Println("ELRS to Joystick emulation")

	port := flag.String("p", "COM10", "serial port to use")
	baudrate := flag.Int("hz", 425000, "baudrate, 425000 by default")
	timeout := flag.Duration("t", 1*time.Second, "timeout to use")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	instance := crsf.New(*port, *baudrate, *timeout)
	err := instance.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if !vjoy.Available() {
		fmt.Println("No joystick available")
		os.Exit(1)
	}
	joystick := getJoystick()

	// Goroutine to handle signals
	go func() {
		<-sigChan

		// Kill CRSF parser:
		err := instance.Close()
		if err != nil {
			fmt.Println(err)
		}

		// Kill joystick:
		joystick.Relinquish()
		fmt.Println("Exiting...")
		os.Exit(0)
	}()

	instance.Parse(func(packet crsf.Packet) {
		const maxValue uint16 = 1811 // On my RC this is the max value for any stick high read.

		if *verbose {
			fmt.Printf("packet: %v\n", packet.Channels)
		}

		joystick.Axis(vjoy.AxisX).Seti(scaleToJoystick(packet.Channels[3], maxValue))
		joystick.Axis(vjoy.AxisY).Seti(scaleToJoystick(packet.Channels[2], maxValue))
		joystick.Axis(vjoy.AxisRX).Seti(scaleToJoystick(packet.Channels[0], maxValue))
		joystick.Axis(vjoy.AxisRY).Seti(scaleToJoystick(packet.Channels[1], maxValue))

		if packet.Channels[4] > 1500 {
			joystick.Button(0).Set(true)
		} else {
			joystick.Button(0).Set(false)
		}

		if packet.Channels[5] > 1500 {
			joystick.Button(1).Set(true)
		} else {
			joystick.Button(1).Set(false)
		}

		if packet.Channels[6] > 1500 {
			joystick.Button(2).Set(true)
		} else {
			joystick.Button(2).Set(false)
		}

		if packet.Channels[7] > 1500 {
			joystick.Button(3).Set(true)
		} else {
			joystick.Button(3).Set(false)
		}

		joystick.Update()
	})
}
