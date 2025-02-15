package crsf

import (
	"testing"

	packet "github.com/konradit/crsf/pkg/crsfpacket"
	"github.com/stretchr/testify/require"
)

func TestInput(t *testing.T) {
	t.Run("Entire packet", func(t *testing.T) {
		input := []byte{200, 24, 22, 219, 195, 94, 46, 190, 7, 112, 240, 129, 15, 224, 224, 3, 31, 248, 40, 8, 0, 0, 76, 124, 226, 193}
		want := packet.ChannelsMap{987, 984, 185, 991, 1792, 992, 992, 1792, 992, 992, 992, 1044, 0, 0, 1811, 1811}

		got := parsePacket(input)

		require.NotNil(t, got)
		require.NotEmpty(t, got)
		notPtr := *got
		require.Equal(t, notPtr, want)
	})

	t.Run("Broken read", func(t *testing.T) {
		packet := []byte{24, 22, 219, 187, 222, 45, 190, 7, 112, 240, 129, 15, 224, 224, 3, 31, 248, 40, 8, 0, 0, 76, 124, 226, 95}

		got := parsePacket(packet)
		require.Nil(t, got)
	})

	t.Run("Append reads to input and successfully parse", func(t *testing.T) {
		data1 := []byte{24, 22, 219, 195, 94, 46, 190, 7, 112, 240, 129, 15, 224, 224, 3, 31, 248, 40, 8, 0, 0, 76, 124, 226, 193}
		data2 := append(data1, []byte{200}...)
		data3 := append(data2, []byte{24, 22, 219, 195, 222, 45, 190, 7, 112, 240, 129, 15, 224, 224, 3, 31, 248, 40, 8, 0, 0, 76, 124, 226, 255}...)
		want := packet.ChannelsMap{987, 984, 183, 991, 1792, 992, 992, 1792, 992, 992, 992, 1044, 0, 0, 1811, 1811}

		got := parsePacket(data1) // no end byte
		require.Nil(t, got)

		got = parsePacket(data2) // no end byte
		require.Nil(t, got)

		got = parsePacket(data3) // can be parsed
		require.NotNil(t, got)
		require.NotEmpty(t, got)
		notPtr := *got
		require.Equal(t, notPtr, want)
	})
}
