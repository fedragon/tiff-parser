package tiff

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestParseEndianness(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte
		order binary.ByteOrder
		err   bool
	}{
		{
			name:  "IntelByteOrder",
			input: []byte{0x49, 0x49},
			order: binary.LittleEndian,
			err:   false,
		},
		{
			name:  "MotorolaByteOrder",
			input: []byte{0x4D, 0x4D},
			order: binary.BigEndian,
			err:   false,
		},
		{
			name:  "UnknownByteOrder",
			input: []byte{0x34, 0x4D},
			order: nil,
			err:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order, err := readEndianness(tc.input)
			if tc.err && err == nil {
				t.Error("expected error, but got none")
			}
			if !tc.err && err != nil {
				t.Error(err)
			}
			if order != tc.order {
				t.Errorf("expected order %v, but got %v", tc.order, order)
			}
		})
	}
}

func Test_ParseMagicNumber(t *testing.T) {
	testCases := []struct {
		name      string
		byteOrder binary.ByteOrder
		input     []byte
		err       bool
	}{
		{
			name:      "TiffMagicNumberBigEndian",
			byteOrder: binary.BigEndian,
			input:     []byte{0x00, 0x2A},
			err:       false,
		},
		{
			name:      "TiffMagicNumberLittleEndian",
			byteOrder: binary.LittleEndian,
			input:     []byte{0x2A, 0x00},
			err:       false,
		},
		{
			name:      "OrfMagicNumberBigEndian",
			byteOrder: binary.BigEndian,
			input:     []byte{0x4F, 0x52},
			err:       false,
		},
		{
			name:      "OrfMagicNumberLittleEndian",
			byteOrder: binary.LittleEndian,
			input:     []byte{0x52, 0x4F},
			err:       false,
		},
		{
			name:      "UnknownMagicNumber",
			byteOrder: binary.BigEndian,
			input:     []byte{0x34, 0x12},
			err:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMagicNumber(tc.byteOrder, tc.input)
			if tc.err && err == nil {
				t.Error("expected error, but got none")
			} else if !tc.err && err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParseDateTimeOriginal_CR2(t *testing.T) {
	r, err := os.Open("../test/data/image.cr2")
	defer r.Close()
	assert.NoError(t, err)

	p, err := NewParser(r)
	if assert.NoError(t, err) {
		date, found, err := p.ParseOriginalDatetime()
		if assert.NoError(t, err) {
			assert.True(t, found)
			assert.Equal(t, time.Date(2021, 11, 19, 12, 21, 10, 0, time.UTC), date)
		}
	}
}

func TestParseDateTimeOriginal_ORF(t *testing.T) {
	r, err := os.Open("../test/data/image.orf")
	defer r.Close()
	assert.NoError(t, err)

	p, err := NewParser(r)
	if assert.NoError(t, err) {
		date, found, err := p.ParseOriginalDatetime()
		if assert.NoError(t, err) {
			assert.True(t, found)
			assert.Equal(t, time.Date(2016, 8, 12, 13, 32, 54, 0, time.UTC), date)
		}
	}
}
