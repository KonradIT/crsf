package crsf

import (
	"errors"
)

var ErrDataTooShort = errors.New("data too short")

type CRSFHeader struct {
	SyncByte    byte
	FrameLength byte
	Type        byte
}

type CRSFFrame struct {
	Header  CRSFHeader
	Payload []byte
	CRC     byte
}

func parseCRSFHeader(data []byte) (CRSFHeader, error) {
	if len(data) < 3 {
		return CRSFHeader{}, ErrDataTooShort
	}

	header := CRSFHeader{
		SyncByte:    data[0],
		FrameLength: data[1],
		Type:        data[2],
	}

	return header, nil
}

func parseCRSFFrame(data []byte) (CRSFFrame, error) {
	if len(data) < 4 {
		return CRSFFrame{}, ErrDataTooShort
	}

	header, err := parseCRSFHeader(data)
	if err != nil {
		return CRSFFrame{}, err
	}

	payloadStart := 3

	payloadLength := int(header.FrameLength) - 2
	if len(data) < payloadStart+payloadLength+1 {
		return CRSFFrame{}, ErrDataTooShort
	}

	payload := data[payloadStart : payloadStart+payloadLength]
	crc := data[payloadStart+payloadLength]

	return CRSFFrame{
		Header:  header,
		Payload: payload,
		CRC:     crc,
	}, nil
}

func unpackChannels(data []byte) channelsMap {
	var channels channelsMap
	bitOffset := 0

	for i := range 16 {
		// Calculate the byte position
		byteIndex := bitOffset / 8
		bitIndex := bitOffset % 8

		// Read up to 3 bytes since an 11-bit value might span across them
		value := uint32(data[byteIndex]) | uint32(data[byteIndex+1])<<8
		if bitIndex > 5 { // If we need bits from the third byte
			value |= uint32(data[byteIndex+2]) << 16
		}

		// Bit shift right (prepend 0s to remove the bits read)
		value >>= bitIndex

		// 0 11111111111
		value &= 0x7FF

		channels[i] = uint16(value)
		bitOffset += 11
	}

	return channels
}
