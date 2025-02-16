package crsfpacket

import (
	"encoding/json"
	"fmt"
)

type PacketType byte

const (
	SyncByte = 0xc8

	// https://github.com/crsf-wg/crsf/wiki/Packet-Types
	FrameChannelsPacked PacketType = 0x16 // 0x16 / 22. rc channels packed type.
	FrameLinkStats      PacketType = 0x14 // 0x14 / 20. Link stats.
)

func (p PacketType) String() string {
	switch p {
	case FrameChannelsPacked:
		return "FrameChannelsPacked"
	case FrameLinkStats:
		return "FrameLinkStats"
	default:
		return fmt.Sprintf("PacketType(%d)", p)
	}
}

func (p PacketType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}
