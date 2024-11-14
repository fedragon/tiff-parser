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
	"github.com/fedragon/tiff-parser/tiff/entry"
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

func TestParse_CR2(t *testing.T) {
	p, err := NewParser(bytes.NewReader(cr2Image))
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

	thumbnail, err := p.ReadThumbnail()
	assert.NoError(t, err)
	assert.Greater(t, len(thumbnail), 1)
	assert.NoError(t, os.WriteFile(path.Join(t.TempDir(), "thumbnail.jpeg"), thumbnail, 0x644))
}

func TestParse_ORF(t *testing.T) {
	p, err := NewParser(bytes.NewReader(orfImage))
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

func TestParser_ReadUints16(t *testing.T) {
	type fields struct {
		reader io.ReadSeeker
	}
	type args struct {
		entry entry.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []uint16
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"returns an error when data type != 3",
			fields{nil},
			args{
				entry.Entry{
					ID:       0,
					DataType: 1,
					Length:   1,
					Value:    0,
				},
			},
			nil,
			assert.Error,
		},
		{
			"returns the entry value when length == 1",
			fields{nil},
			args{
				entry.Entry{
					ID:       0,
					DataType: 3,
					Length:   1,
					Value:    111,
				},
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
				entry.Entry{
					ID:       0,
					DataType: 3,
					Length:   2,
					Value:    0,
				},
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
			got, err := p.ReadUints16(tt.args.entry)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadUints16(%v)", tt.args.entry)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ReadUints16(%v)", tt.args.entry)
		})
	}
}

func TestParser_ReadUints32(t *testing.T) {
	type fields struct {
		reader io.ReadSeeker
	}
	type args struct {
		entry entry.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []uint32
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"returns an error when data type != 4",
			fields{nil},
			args{
				entry.Entry{
					ID:       0,
					DataType: 1,
					Length:   1,
					Value:    0,
				},
			},
			nil,
			assert.Error,
		},
		{
			"returns the entry value when length == 1",
			fields{nil},
			args{
				entry.Entry{
					ID:       0,
					DataType: 4,
					Length:   1,
					Value:    111,
				},
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
				entry.Entry{
					ID:       0,
					DataType: 4,
					Length:   2,
					Value:    0,
				},
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
			got, err := p.ReadUints32(tt.args.entry)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadUints32(%v)", tt.args.entry)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ReadUints32(%v)", tt.args.entry)
		})
	}
}

func TestParser_ReadURational(t *testing.T) {
	type fields struct {
		reader io.ReadSeeker
	}
	type args struct {
		entry entry.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantNum uint32
		wantDen uint32
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"returns an error when data type != 5",
			fields{nil},
			args{
				entry.Entry{
					ID:       0,
					DataType: 1,
					Length:   1,
					Value:    0,
				},
			},
			0,
			0,
			assert.Error,
		},
		{
			"returns numerator and denominator",
			fields{
				test.NewBytesReadSeeker().WithUints32(111, 222),
			},
			args{
				entry.Entry{
					ID:       0,
					DataType: 5,
					Length:   1,
					Value:    0,
				},
			},
			111,
			222,
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				reader:    tt.fields.reader,
				byteOrder: binary.LittleEndian,
			}
			gotNum, gotDen, err := p.ReadURational(tt.args.entry)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadURational(%v)", tt.args.entry)) {
				return
			}
			assert.Equalf(t, tt.wantNum, gotNum, "ReadURational(%v)", tt.args.entry)
			assert.Equalf(t, tt.wantDen, gotDen, "ReadURational(%v)", tt.args.entry)
		})
	}
}

func TestParser_ReadString(t *testing.T) {
	type fields struct {
		reader io.ReadSeeker
	}
	type args struct {
		entry entry.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"returns an error when data type != 2",
			fields{nil},
			args{
				entry.Entry{
					ID:       0,
					DataType: 1,
					Length:   1,
					Value:    0,
				},
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
				entry.Entry{
					ID:       0,
					DataType: 2,
					Length:   4,
					Value:    0,
				},
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
			got, err := p.ReadString(tt.args.entry)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadString(%v)", tt.args.entry)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ReadString(%v)", tt.args.entry)
		})
	}
}

func TestPrintEntries(t *testing.T) {
	p, err := NewParser(bytes.NewReader(orfImage))
	assert.NoError(t, err)

	p.PrintEntries(p.firstIFDOffset)
}
