package rtmp

import (
	"testing"
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
