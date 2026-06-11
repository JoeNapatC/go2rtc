package rtmp

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

// https://github.com/AlexxIT/go2rtc - panic: index out of range [15] with length 15
// encodeFLV must not panic when the buffer is exactly header-sized (15 bytes),
// which happens when an RTMP message arrives with an empty payload.
func TestEncodeFLVEmptyPayload(t *testing.T) {
	b := make([]byte, 4+11)
	encodeFLV(b, 8, 0x01020304, nil)

	if b[4] != 8 {
		t.Errorf("wrong msgType: %d", b[4])
	}
	if b[5] != 0 || b[6] != 0 || b[7] != 0 {
		t.Errorf("wrong payload size: % x", b[5:8])
	}
	if b[8] != 2 || b[9] != 3 || b[10] != 4 || b[11] != 1 {
		t.Errorf("wrong timestamp: % x", b[8:12])
	}
}

func TestEncodeFLVWithPayload(t *testing.T) {
	payload := []byte{0xAA, 0xBB, 0xCC}
	b := make([]byte, 4+11+len(payload))
	encodeFLV(b, 9, 0, payload)

	if b[4] != 9 {
		t.Errorf("wrong msgType: %d", b[4])
	}
	if b[7] != 3 {
		t.Errorf("wrong payload size: %d", b[7])
	}
	if b[15] != 0xAA || b[16] != 0xBB || b[17] != 0xCC {
		t.Errorf("wrong payload: % x", b[15:])
	}
}

// Encoders (ex. drones) can send Set Chunk Size after the publish command.
// Read should apply it instead of converting it to a FLV tag,
// otherwise all following chunks will be parsed at a wrong size.
func TestReadSetPacketSize(t *testing.T) {
	var raw []byte

	// Set Chunk Size = 8 (chunkID=2, header type 0)
	raw = append(raw, 0x02, 0, 0, 0, 0, 0, 4, TypeSetPacketSize, 0, 0, 0, 0)
	raw = append(raw, 0, 0, 0, 8)

	// video message 16 bytes (chunkID=4), split into two chunks of 8
	payload := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	raw = append(raw, 0x04, 0, 0, 0, 0, 0, 16, TypeVideo, 0, 0, 0, 0)
	raw = append(raw, payload[:8]...)
	raw = append(raw, 3<<6|0x04) // continuation chunk
	raw = append(raw, payload[8:]...)

	c := &Conn{
		rd:           bytes.NewReader(raw),
		chunks:       map[byte]*chunk{},
		rdPacketSize: 128,
	}

	b := make([]byte, 64)
	n, err := c.Read(b)
	require.Nil(t, err)
	require.Equal(t, 4+11+16, n)
	require.Equal(t, byte(TypeVideo), b[4])
	require.Equal(t, payload, b[15:n])
	require.Equal(t, uint32(8), c.rdPacketSize)
}

// Server should acknowledge received bytes when remote requested it
// via Window Acknowledgement Size
func TestSendAcks(t *testing.T) {
	wr := &bytes.Buffer{}
	c := &Conn{
		cnt:          &countReader{n: 5000},
		rdAckWindow:  4096,
		wr:           wr,
		wrPacketSize: 4096,
	}

	c.sendAcks()
	b := wr.Bytes()
	require.Len(t, b, 12+4)
	require.Equal(t, byte(TypeAcknowledgement), b[7])
	require.Equal(t, uint32(5000), binary.BigEndian.Uint32(b[12:]))

	// no new ack until the next window
	wr.Reset()
	c.sendAcks()
	require.Empty(t, wr.Bytes())
}
