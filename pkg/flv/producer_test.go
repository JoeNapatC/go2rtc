package flv

import (
	"bytes"
	"testing"

	"github.com/AlexxIT/go2rtc/pkg/core"
	"github.com/stretchr/testify/require"
)

func appendTag(b []byte, tagType byte, payload []byte) []byte {
	b = append(b, 0, 0, 0, 0) // previous tag size
	b = append(b, tagType)
	b = append(b, byte(len(payload)>>16), byte(len(payload)>>8), byte(len(payload)))
	b = append(b, 0, 0, 0, 0) // timestamp + extended
	b = append(b, 0, 0, 0)    // stream ID
	return append(b, payload...)
}

// High bitrate video without audio and metadata (ex. drone stream) can fill
// the probe buffer before the probe timeout. Probe should go with the media
// it has instead of failing with "probe reader overflow".
func TestProbeOverflowVideoOnly(t *testing.T) {
	sps := []byte{0x67, 0x42, 0xC0, 0x1E, 0xD9, 0x00, 0x40}
	pps := []byte{0x68, 0xCE, 0x38, 0x80}

	avcC := []byte{1, 0x42, 0xC0, 0x1E, 0xFF, 0xE1, 0x00, byte(len(sps))}
	avcC = append(avcC, sps...)
	avcC = append(avcC, 0x01, 0x00, byte(len(pps)))
	avcC = append(avcC, pps...)

	b := []byte{'F', 'L', 'V', 1, 0, 0, 0, 0, 9}

	// AVC sequence header (key frame + AVC codec, packet type header)
	hdr := append([]byte{0x17, PacketTypeAVCHeader, 0, 0, 0}, avcC...)
	b = appendTag(b, TagVideo, hdr)

	// big inter frames (inter frame + AVC codec, packet type NALU)
	frame := make([]byte, 0xFFFF)
	frame[0] = 0x27
	frame[1] = PacketTypeAVCNALU
	for len(b) < core.ProbeSize*2 {
		b = appendTag(b, TagVideo, frame)
	}

	prod, err := Open(bytes.NewReader(b))
	require.Nil(t, err)
	require.Len(t, prod.Medias, 1)
	require.Equal(t, core.KindVideo, prod.Medias[0].Kind)
}

// Tags with too short payload (ex. drones can send empty media messages)
// should be skipped without panic
func TestProbeShortPayload(t *testing.T) {
	b := []byte{'F', 'L', 'V', 1, 0, 0, 0, 0, 9}
	b = appendTag(b, TagVideo, nil)
	b = appendTag(b, TagAudio, []byte{0xAF})

	_, err := Open(bytes.NewReader(b))
	require.NotNil(t, err) // EOF, but no panic and no media
}
