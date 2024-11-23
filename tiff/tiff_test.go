package tiff

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/fedragon/tiff-parser/test"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/image.cr2
var cr2Image []byte

//go:embed testdata/image.orf
var orfImage []byte

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

func TestParser_readUints16(t *testing.T) {
	type fields struct {
		reader io.ReadSeeker
	}
	type args struct {
		length uint32
		offset uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []uint16
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"returns an error when seek fails",
			fields{
				test.NewBytesReadSeeker(),
			},
			args{
				length: 1,
				offset: 0,
			},
			nil,
			assert.Error,
		},
		{
			"returns the entry value when length == 1",
			fields{
				test.NewBytesReadSeeker().WithUints16(111),
			},
			args{
				length: 1,
				offset: 0,
			},
			[]uint16{111},
			assert.NoError,
		},
		{
			"returns all values when length > 1",
			fields{
				test.NewBytesReadSeeker().WithUints16(111, 222),
			},
			args{
				length: 2,
				offset: 0,
			},
			[]uint16{111, 222},
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				reader:    tt.fields.reader,
				byteOrder: binary.LittleEndian,
			}
			got, err := p.readUints16(tt.args.length, tt.args.offset)
			if !tt.wantErr(t, err, fmt.Sprintf("readUints16(%v, %v)", tt.args.length, tt.args.offset)) {
				return
			}
			assert.Equalf(t, tt.want, got, "readUints16(%v, %v)", tt.args.length, tt.args.offset)
		})
	}
}

func TestParser_readUints32(t *testing.T) {
	type fields struct {
		reader io.ReadSeeker
	}
	type args struct {
		length uint32
		offset uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []uint32
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"returns an error when seek fails",
			fields{
				test.NewBytesReadSeeker(),
			},
			args{
				length: 1,
				offset: 0,
			},
			nil,
			assert.Error,
		},
		{
			"returns the entry value when length == 1",
			fields{
				test.NewBytesReadSeeker().WithUints32(111),
			},
			args{
				length: 1,
				offset: 0,
			},
			[]uint32{111},
			assert.NoError,
		},
		{
			"returns all values when length > 1",
			fields{
				test.NewBytesReadSeeker().WithUints32(111, 222),
			},
			args{
				length: 2,
				offset: 0,
			},
			[]uint32{111, 222},
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				reader:    tt.fields.reader,
				byteOrder: binary.LittleEndian,
			}
			got, err := p.readUints32(tt.args.length, tt.args.offset)
			if !tt.wantErr(t, err, fmt.Sprintf("readUints32(%v, %v)", tt.args.length, tt.args.offset)) {
				return
			}
			assert.Equalf(t, tt.want, got, "readUints32(%v, %v)", tt.args.length, tt.args.offset)
		})
	}
}

func TestParser_readURational(t *testing.T) {
	type fields struct {
		reader io.ReadSeeker
	}
	type args struct {
		offset uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    URational
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"returns an error when seek fails",
			fields{
				test.NewBytesReadSeeker(),
			},
			args{
				offset: 0,
			},
			URational{},
			assert.Error,
		},
		{
			"returns numerator and denominator",
			fields{
				test.NewBytesReadSeeker().WithUints32(111, 222),
			},
			args{
				offset: 0,
			},
			URational{111, 222},
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				reader:    tt.fields.reader,
				byteOrder: binary.LittleEndian,
			}
			got, err := p.readURational(tt.args.offset)
			if !tt.wantErr(t, err, fmt.Sprintf("readURational(%v)", tt.args.offset)) {
				return
			}
			assert.Equalf(t, tt.want, got, "readURational(%v)", tt.args.offset)
		})
	}
}

func TestParser_readString(t *testing.T) {
	type fields struct {
		reader io.ReadSeeker
	}
	type args struct {
		length uint32
		offset uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"returns an error when seek fails",
			fields{
				test.NewBytesReadSeeker(),
			},
			args{
				length: 1,
				offset: 0,
			},
			"",
			assert.Error,
		},
		{
			"returns string",
			fields{
				test.NewBytesReadSeeker().WithString("abc\000"),
			},
			args{
				length: 4,
				offset: 0,
			},
			"abc",
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				reader:    tt.fields.reader,
				byteOrder: binary.LittleEndian,
			}
			got, err := p.readString(tt.args.length, tt.args.offset)
			if !tt.wantErr(t, err, fmt.Sprintf("readString(%v, %v)", tt.args.length, tt.args.offset)) {
				return
			}
			assert.Equalf(t, tt.want, got, "readString(%v, %v)", tt.args.length, tt.args.offset)
		})
	}
}

func TestPrintEntries(t *testing.T) {
	p, err := NewParser(bytes.NewReader(cr2Image))
	assert.NoError(t, err)
	assert.NoError(t, p.PrintEntries())
}

func TestParse_CR2(t *testing.T) {
	p, err := NewParser(bytes.NewReader(cr2Image))
	assert.NoError(t, err)

	entries, err := p.Parse(ImageWidth, ImageHeight, BitsPerSample, Make, DateTimeOriginal, ExposureTime)

	assert.NoError(t, err)
	assert.NotEmpty(t, entries)

	width := entries[ImageWidth].Value.Uint16
	assert.NotNil(t, width)
	assert.EqualValues(t, 5184, *width)

	height := entries[ImageHeight].Value.Uint16
	assert.NotNil(t, height)
	assert.EqualValues(t, 3456, *height)

	bitsPerSample := entries[BitsPerSample].Value.Uints16
	assert.ElementsMatch(t, [3]uint16{8, 8, 8}, bitsPerSample)

	make_ := entries[Make].Value.String
	assert.NotNil(t, make_)
	assert.EqualValues(t, "Canon", *make_)

	dateTime := entries[DateTimeOriginal].Value.String
	assert.NotNil(t, dateTime)
	assert.EqualValues(t, "2021:11:19 12:21:10", *dateTime)

	numden := entries[ExposureTime].Value.URational
	assert.NotNil(t, numden)
	assert.EqualValues(t, 1, numden.Numerator)
	assert.EqualValues(t, 40, numden.Denominator)

	thumbnail, err := p.ReadThumbnail()
	assert.NoError(t, err)
	assert.Greater(t, len(thumbnail), 1)
	assert.NoError(t, os.WriteFile(path.Join(t.TempDir(), "thumbnail.jpeg"), thumbnail, 0x644))
}

func TestParse_ORF(t *testing.T) {
	p, err := NewParser(bytes.NewReader(orfImage))
	assert.NoError(t, err)

	entries, err := p.Parse(ImageWidth, ImageHeight, BitsPerSample, Make, DateTimeOriginal, ExposureTime)

	assert.NoError(t, err)
	assert.NotEmpty(t, entries)

	width := entries[ImageWidth].Value.Uint32
	assert.NotNil(t, width)
	assert.EqualValues(t, 4640, *width)

	height := entries[ImageHeight].Value.Uint32
	assert.NotNil(t, height)
	assert.EqualValues(t, 3472, *height)

	fmt.Printf("%v\n", entries[BitsPerSample].Value)

	bitsPerSample := entries[BitsPerSample].Value.Uint16
	assert.NotNil(t, bitsPerSample)
	assert.EqualValues(t, 16, *bitsPerSample)

	make_ := entries[Make].Value.String
	assert.NotNil(t, make_)
	assert.EqualValues(t, "OLYMPUS CORPORATION    ", *make_)

	numden := entries[ExposureTime].Value.URational
	assert.NotNil(t, numden)
	assert.EqualValues(t, 1, numden.Numerator)
	assert.EqualValues(t, 200, numden.Denominator)

	dateTime := entries[DateTimeOriginal].Value.String
	assert.NotNil(t, dateTime)
	assert.EqualValues(t, "2016:08:12 13:32:54", *dateTime)
}
