package test

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type endianness interface {
	binary.ByteOrder
	binary.AppendByteOrder
}

type BytesReadSeeker struct {
	byteOrder endianness
	buffer    []byte
	offset    int64

	io.ReadSeeker
}

func NewBytesReadSeeker() *BytesReadSeeker {
	return &BytesReadSeeker{
		byteOrder: binary.LittleEndian,
		buffer:    make([]byte, 0),
	}
}

func (brs *BytesReadSeeker) WithString(value string) *BytesReadSeeker {
	brs.buffer = append(brs.buffer, []byte(value)...)

	return brs
}

func (brs *BytesReadSeeker) WithUints16(values ...uint16) *BytesReadSeeker {
	for _, value := range values {
		brs.buffer = brs.byteOrder.AppendUint16(brs.buffer, value)
	}

	return brs
}

func (brs *BytesReadSeeker) WithUints32(values ...uint32) *BytesReadSeeker {
	for _, value := range values {
		brs.buffer = brs.byteOrder.AppendUint32(brs.buffer, value)
	}

	return brs
}

func (brs *BytesReadSeeker) Read(p []byte) (int, error) {
	if p == nil {
		return 0, errors.New("destination cannot be nil")
	}

	if len(p) >= len(brs.buffer) {
		n := copy(p, brs.buffer)
		return n, nil
	}

	n := copy(p, brs.buffer[0:len(p)])
	return n, nil
}

func (brs *BytesReadSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart {
		return 0, errors.New("can only seek from start")
	}

	if offset < 0 {
		return 0, errors.New("negative offset not allowed")
	}

	if offset >= int64(len(brs.buffer)) {
		return 0, fmt.Errorf("offset %d exceeds buffer length %d", offset, len(brs.buffer))
	}

	brs.offset = offset

	return brs.offset, nil
}
