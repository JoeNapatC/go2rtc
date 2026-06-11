package core

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadSeeker(t *testing.T) {
	b := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	buf := bytes.NewReader(b)

	rd := NewReadBuffer(buf)
	rd.BufferSize = ProbeSize

	// 1. Read to buffer
	b = make([]byte, 3)
	n, err := rd.Read(b)
	require.Nil(t, err)
	require.Equal(t, []byte{0, 1, 2}, b[:n])

	// 2. Seek to start
	_, err = rd.Seek(0, io.SeekStart)
	require.Nil(t, err)

	// 3. Read from buffer
	b = make([]byte, 2)
	n, err = rd.Read(b)
	require.Nil(t, err)
	require.Equal(t, []byte{0, 1}, b[:n])

	// 4. Read from buffer
	n, err = rd.Read(b)
	require.Nil(t, err)
	require.Equal(t, []byte{2}, b[:n])

	// 5. Read to buffer
	n, err = rd.Read(b)
	require.Nil(t, err)
	require.Equal(t, []byte{3, 4}, b[:n])

	// 6. Seek to start
	_, err = rd.Seek(0, io.SeekStart)
	require.Nil(t, err)

	// 7. Disable buffer
	rd.BufferSize = -1

	// 8. Read from buffer
	b = make([]byte, 10)
	n, err = rd.Read(b)
	require.Nil(t, err)
	require.Equal(t, []byte{0, 1, 2, 3, 4}, b[:n])

	// 9. Direct read
	n, err = rd.Read(b)
	require.Nil(t, err)
	require.Equal(t, []byte{5, 6, 7, 8, 9}, b[:n])

	// 10. Check buffer empty
	require.Nil(t, rd.buf)
}

func TestReadBufferOverflow(t *testing.T) {
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}

	rd := NewReadBuffer(bytes.NewReader(data))
	rd.BufferSize = 10

	// 1. Read to buffer
	b := make([]byte, 7)
	n, err := rd.Read(b)
	require.Nil(t, err)
	require.Equal(t, 7, n)

	// 2. Buffer can exceed BufferSize by one read
	n, err = rd.Read(b)
	require.Nil(t, err)
	require.Equal(t, 7, n)

	// 3. Overflow without reading from the source
	_, err = rd.Read(b)
	require.ErrorIs(t, err, ErrProbeOverflow)

	// 4. After Reset all data should be available (no bytes lost)
	rd.Reset()
	all, err := io.ReadAll(rd)
	require.Nil(t, err)
	require.Equal(t, data, all)
}
