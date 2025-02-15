package crsf

import (
	"errors"
	"fmt"
	"time"

	"go.bug.st/serial"
)

var ErrConnNotInitialized = errors.New("serial: conn not initialized")
var ErrWhenReading = func(err error) error {
	return fmt.Errorf("Parse: failed to read from serial: %w", err)
}

type CRSFParse struct {
	Device   string
	Baudrate int
	Timeout  time.Duration

	serialConn serial.Port
}

// first 4 are corresponding to throttle, yaw, pitch, roll.
// packet: [987 984 1386 991 1792 992 992 1792 992 992 992 1044 0 0 1811 1811]
// 0 => right stick left/right
// 1 => right stick up/down
// 2 => left stick up/down
// 3 => left stick left/right

// todo: ditch unused channels.
type channelsMap [16]uint16

type Packet struct {
	Channels channelsMap
}

type PacketCallback func(packet Packet)

// New() creates a new Parser instance.
// device: the serial device to connect to.
// baudrate: Use 425000 because that's the baudrate for the ELRS modules.
// timeout: the timeout to use in duration (eg: 2*time.Second)
func New(device string, baudrate int, timeout time.Duration) *CRSFParse {
	return &CRSFParse{
		Device:   device,
		Baudrate: baudrate,
		Timeout:  timeout,
	}
}

// Start() opens the serial connection and starts parsing the packets.
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

// Close() closes the serial connection.
func (c *CRSFParse) Close() error {
	if c.serialConn == nil {
		return ErrConnNotInitialized
	}

	return c.serialConn.Close()
}

const (
	syncByte = 0xc8
)

func parsePacket(data []byte) *channelsMap {
	for len(data) > 0 {
		if data[0] == syncByte { // sync byte.
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
			return &channels
		}
		data = data[1:]
	}
	return nil
}

// Parse() reads from serial connection, attempts to parse data and returns the channels in the callback function.
func (c *CRSFParse) Parse(callback PacketCallback) error {
	if c.serialConn == nil {
		return ErrConnNotInitialized
	}

	maxSize := (22 * 4)

	buf := make([]byte, 0, maxSize*2) // 2x maxSize to accommodate for partial reads.
	for {
		// Read in small chunks like Python
		temp := make([]byte, maxSize)
		n, err := c.serialConn.Read(temp)
		if err != nil {
			return ErrWhenReading(err)
		}

		buf = append(buf, temp[:n]...)

		// Try to parse
		channels := parsePacket(buf)

		if channels != nil {
			callback(Packet{Channels: *channels})
			buf = buf[:0]
		}

		if len(buf) > maxSize {
			buf = buf[len(buf)-maxSize:]
		}
	}
}
