package crsfpacket

import (
	"errors"
)

var ErrDataTooShort = errors.New("data too short")

// first 4 are corresponding to throttle, yaw, pitch, roll.
// packet: [987 984 1386 991 1792 992 992 1792 992 992 992 1044 0 0 1811 1811]
// 0 => right stick left/right
// 1 => right stick up/down
// 2 => left stick up/down
// 3 => left stick left/right

// todo: ditch unused channels.
type ChannelsMap [16]uint16

type Packet struct {
	Channels ChannelsMap `json:"channels"`
}

type Header struct {
	SyncByte    byte `json:"sync_byte"`
	FrameLength byte `json:"frame_length"`
	Type        byte `json:"type"`
}

type Frame struct {
	Header  Header `json:"header"`
	Payload []byte `json:"payload"`
	CRC     byte   `json:"crc"`
}

func parseHeader(data []byte) (Header, error) {
	if len(data) < 3 {
		return Header{}, ErrDataTooShort
	}

	header := Header{
		SyncByte:    data[0],
		FrameLength: data[1],
		Type:        data[2],
	}

	return header, nil
}

func ParseFrame(data []byte) (Frame, error) {
	if len(data) < 4 {
		return Frame{}, ErrDataTooShort
	}

	header, err := parseHeader(data)
	if err != nil {
		return Frame{}, err
	}

	payloadStart := 3

	payloadLength := int(header.FrameLength) - 2
	if len(data) < payloadStart+payloadLength+1 {
		return Frame{}, ErrDataTooShort
	}

	payload := data[payloadStart : payloadStart+payloadLength]
	crc := data[payloadStart+payloadLength]

	return Frame{
		Header:  header,
		Payload: payload,
		CRC:     crc,
	}, nil
}

func UnpackChannels(data []byte) ChannelsMap {
	var channels ChannelsMap
	bitOffset := 0

	for i := range 16 {
		// Calculate the byte position
		byteIndex := bitOffset / 8
		bitIndex := bitOffset % 8

		// Read up to 3 bytes since an 11-bit value might span across them
		// Using uint32 because data is too large (11 bits), uint16 will be casted later.
		value := uint32(data[byteIndex]) | uint32(data[byteIndex+1])<<8
		if bitIndex > 5 { // If we need bits from the third byte
			value |= uint32(data[byteIndex+2]) << 16
		}

		// Bit shift right (prepend 0s to remove the bits read)
		value >>= bitIndex

		// 0 11111111111
		value &= 0x7FF

		channels[i] = uint16(value) //nolint:G115 // Max 2048 here. Not an issue.
		bitOffset += 11
	}

	return channels
}
