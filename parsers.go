package crsf

import "fmt"

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
		return CRSFHeader{}, fmt.Errorf("data too short for header")
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
		return CRSFFrame{}, fmt.Errorf("data too short for frame")
	}

	header, err := parseCRSFHeader(data)
	if err != nil {
		return CRSFFrame{}, err
	}

	payloadStart := 3

	payloadLength := int(header.FrameLength) - 2
	if len(data) < payloadStart+payloadLength+1 {
		return CRSFFrame{}, fmt.Errorf("data too short for payload")
	}

	payload := data[payloadStart : payloadStart+payloadLength]
	crc := data[payloadStart+payloadLength]

	return CRSFFrame{
		Header:  header,
		Payload: payload,
		CRC:     crc,
	}, nil
}

func unpackChannels(data []byte) []uint16 {
	var channels []uint16
	bitOffset := 0

	for i := 0; i < 16; i++ {
		// Calculate the byte and bit position
		byteIndex := bitOffset / 8
		bitIndex := bitOffset % 8

		// Extract 11 bits for each channel
		value := uint16(data[byteIndex]) | uint16(data[byteIndex+1])<<8
		value >>= bitIndex
		value &= 0x7FF // Mask to 11 bits

		channels = append(channels, value)
		bitOffset += 11
	}

	return channels
}
