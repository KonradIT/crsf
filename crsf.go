package crsf

import (
	"errors"
	"fmt"
	"time"

	"go.bug.st/serial"
)

type CRSFParse struct {
	Device   string
	Baudrate int
	Timeout  time.Duration

	serialConn serial.Port
}

// first 4 are corresponding to throttle, yaw, pitch, roll.
type channelsMap [16]uint16

type Packet struct {
	Channels []uint16
}

type PacketCallback func(packet Packet)

func New(device string, baudrate int, timeout time.Duration) *CRSFParse {
	return &CRSFParse{
		Device:   device,
		Baudrate: baudrate,
		Timeout:  timeout,
	}
}

func (c *CRSFParse) Start() error {
	mode := &serial.Mode{
		BaudRate: c.Baudrate,
	}

	conn, err := serial.Open(c.Device, mode)
	if err != nil {
		return fmt.Errorf("Start: failed to open serial conn: %w", err)
	}

	c.serialConn = conn

	return nil
}

func (c *CRSFParse) Close() error {
	return c.serialConn.Close()
}

const (
	SyncByte = 0xc8
)

func parsePacket(data []byte) []uint16 {
	for len(data) > 0 {
		if data[0] == SyncByte { // sync byte.
			if len(data) < 22 {
				return nil
			}
			// validate the size
			packetSize := data[1] - 2
			packetType := data[2]

			expectedSize := byte(22)   // rc channels size.
			expectedType := byte(0x16) // rc channels type.

			if packetSize != expectedSize || packetType != expectedType {
				return nil
			}

			stripped := data[:packetSize+4]

			content, err := parseCRSFFrame(stripped)
			if err != nil {
				return nil
			}

			channels := unpackChannels(content.Payload)
			return channels
		}
		data = data[1:]
	}
	return nil
}

func (c *CRSFParse) Parse(callback PacketCallback) error {
	if c.serialConn == nil {
		return errors.New("Serial: conn not initialized")
	}

	maxSize := (22 * 4)

	buf := make([]byte, 0, maxSize*2) // 2x maxSize because maybe its partially read.
	for {
		// Read in small chunks like Python
		temp := make([]byte, maxSize)
		n, err := c.serialConn.Read(temp)
		if err != nil {
			return fmt.Errorf("Parse: failed to read from serial: %w", err)
		}

		buf = append(buf, temp[:n]...)

		// Try to parse
		channels := parsePacket(buf)

		if channels != nil {
			callback(Packet{Channels: channels})
			buf = buf[:0]
		}

		if len(buf) > maxSize {
			buf = buf[len(buf)-maxSize:]
		}
	}
}
