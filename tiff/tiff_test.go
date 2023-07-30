package tiff

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/fedragon/tiff-parser/tiff/entry"
	"github.com/stretchr/testify/assert"
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

func TestParse_CR2(t *testing.T) {
	r, err := os.Open("../test/data/image.cr2")
	assert.NoError(t, err)
	defer r.Close()

	p, err := NewParser(r)
	assert.NoError(t, err)

	entries, err := p.Parse(entry.ImageWidth, entry.ImageHeight, entry.BitsPerSample, entry.Make, entry.DateTimeOriginal, entry.ExposureTime)

	assert.NoError(t, err)
	assert.NotEmpty(t, entries)

	width, err := p.ReadUint16(entries[entry.ImageWidth])
	assert.NoError(t, err)
	assert.EqualValues(t, 5184, width)

	height, err := p.ReadUint16(entries[entry.ImageHeight])
	assert.NoError(t, err)
	assert.EqualValues(t, 3456, height)

	bitsPerSample, err := p.ReadUints16(entries[entry.BitsPerSample])
	assert.NoError(t, err)
	assert.ElementsMatch(t, [3]uint16{8, 8, 8}, bitsPerSample)

	make_, err := p.ReadString(entries[entry.Make])
	assert.NoError(t, err)
	assert.EqualValues(t, "Canon", make_)

	dateTime, err := p.ReadString(entries[entry.DateTimeOriginal])
	assert.NoError(t, err)
	assert.EqualValues(t, "2021:11:19 12:21:10", dateTime)

	num, den, err := p.ReadURational(entries[entry.ExposureTime])
	assert.NoError(t, err)
	assert.EqualValues(t, 1, num)
	assert.EqualValues(t, 40, den)
}

func TestParse_ORF(t *testing.T) {
	r, err := os.Open("../test/data/image.orf")
	assert.NoError(t, err)
	defer r.Close()

	p, err := NewParser(r)
	assert.NoError(t, err)

	entries, err := p.Parse(entry.ImageWidth, entry.ImageHeight, entry.BitsPerSample, entry.Make, entry.DateTimeOriginal, entry.ExposureTime)

	assert.NoError(t, err)
	assert.NotEmpty(t, entries)

	width, err := p.ReadUint32(entries[entry.ImageWidth])
	assert.NoError(t, err)
	assert.EqualValues(t, 4640, width, int(width))

	height, err := p.ReadUint32(entries[entry.ImageHeight])
	assert.NoError(t, err)
	assert.EqualValues(t, 3472, height, int(height))

	bitsPerSample, err := p.ReadUint16(entries[entry.BitsPerSample])
	assert.NoError(t, err)
	assert.EqualValues(t, 16, bitsPerSample)

	make_, err := p.ReadString(entries[entry.Make])
	assert.NoError(t, err)
	assert.EqualValues(t, "OLYMPUS CORPORATION    ", make_)

	num, den, err := p.ReadURational(entries[entry.ExposureTime])
	assert.NoError(t, err)
	assert.EqualValues(t, 1, num)
	assert.EqualValues(t, 200, den)

	dateTime, err := p.ReadString(entries[entry.DateTimeOriginal])
	assert.NoError(t, err)
	assert.EqualValues(t, "2016:08:12 13:32:54", dateTime)
}
