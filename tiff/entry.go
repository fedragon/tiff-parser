package tiff

import (
	"bytes"
	"errors"
	"fmt"
)

type EntryID uint16
type DataType uint16

const (
	// Length of an IFD entry, in bytes
	EntryLength = 12

	// IFD #0

	ImageWidth    EntryID = 0x100
	ImageHeight   EntryID = 0x101
	BitsPerSample EntryID = 0x102
	Compression   EntryID = 0x103
	Make          EntryID = 0x10f
	Model         EntryID = 0x110
	Exif          EntryID = 0x8769
	GPSInfo       EntryID = 0x8825

	// Exif sub-IFD

	ExposureTime       EntryID = 0x829a
	FNumber            EntryID = 0x829d
	ISO                EntryID = 0x8827
	DateTimeOriginal   EntryID = 0x9003
	OffsetTimeOriginal EntryID = 0x9011

	// GPSInfo sub-IFD

	GPSLatitude  EntryID = 0x0002
	GPSLongitude EntryID = 0x0004

	// Position depends on actual format

	ThumbnailOffset EntryID = 0x0201 // in IFD #1 (PreviewImageStart if in IFD #0)
	ThumbnailLength EntryID = 0x0202 // in IFD #1 (PreviewImageLength if in IFD #0)
)

const (
	DataType_UByte DataType = iota + 1
	DataType_String
	DataType_UShort
	DataType_ULong
	DataType_URational
	DataType_Byte
	DataType_UByte_Sequence
	DataType_Short
	DataType_Long
	DataType_Rational
	DataType_Single_Precision_IEEE_Format
	DataType_Double_Precision_IEEE_Format
)

// Entry represents an IFD entry
type Entry struct {
	ID       EntryID
	DataType DataType
	Length   uint32
	RawValue uint32 // value of the entry or offset to read the value from, depending on DataType and Length
}

func (e Entry) String(s Source) (string, error) {
	dt := "UNKNOWN"
	switch DataType(e.DataType) {
	case DataType_UByte:
		dt = "unsigned byte"
	case DataType_String:
		dt = "string"
	case DataType_UShort:
		dt = "unsigned short 16bits"
	case DataType_ULong:
		dt = "unsigned long 32bits"
	case DataType_URational:
		dt = "unsigned rational"
	case DataType_Byte:
		dt = "signed byte"
	case DataType_UByte_Sequence:
		dt = "unsigned byte sequence"
	case DataType_Short:
		dt = "signed short 16bits"
	case DataType_Long:
		dt = "signed long 32bits"
	case DataType_Rational:
		dt = "signed rational"
	case DataType_Single_Precision_IEEE_Format:
		dt = "single precision (2 bytes) IEEE format"
	case DataType_Double_Precision_IEEE_Format:
		dt = "double precision (4 bytes) IEEE format"
	}

	res, err := e.PrintValue(s)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ID: 0x%X\nDataType: %s\nLength: %d\nValue: %s\n", e.ID, dt, e.Length, res), nil
}

// ReadString reads and returns a string from an IFD entry, trimming its NUL-byte terminator. It returns an error if it cannot read the string.
func (e Entry) ReadString(s Source) (string, error) {
	if e.DataType != DataType_String {
		return "", errors.New("entry is not a string")
	}

	if err := s.SeekFromStart(int64(e.RawValue)); err != nil {
		return "", err
	}

	res := make([]byte, e.Length)
	if err := s.ReadInto(res); err != nil {
		return "", err
	}

	return string(bytes.TrimSuffix(res, []byte{0x0})), nil
}

// ReadUint16 reads and returns a uint16 from an IFD entry. It returns an error if the type doesn't match.
func (e Entry) ReadUint16() (uint16, error) {
	if e.DataType != DataType_UShort {
		return 0, errors.New("entry is not a uint16")
	}

	if e.Length != 1 {
		return 0, errors.New("entry length is not 1")
	}

	return uint16(e.RawValue), nil
}

// ReadUints16 reads and returns a slice of uint16 from an IFD entry. It returns an error if it cannot read the slice.
func (e Entry) ReadUints16(s Source) ([]uint16, error) {
	if e.DataType != DataType_UShort {
		return nil, errors.New("entry is not a []uint16")
	}

	if e.Length == 1 {
		return []uint16{uint16(e.RawValue)}, nil
	}

	res := make([]uint16, e.Length)
	if err := s.SeekFromStart(int64(e.RawValue)); err != nil {
		return nil, err
	}

	size := 2
	buffer := make([]byte, size*int(e.Length))
	if err := s.ReadInto(buffer); err != nil {
		return nil, err
	}

	for i := 0; i < int(e.Length); i++ {
		res[i] = s.Uint16(buffer[i*size : i*size+size])
	}

	return res, nil
}

// ReadUint32 reads and returns a uint32 from an IFD entry. It returns an error if the type doesn't match.
func (e Entry) ReadUint32() (uint32, error) {
	if e.DataType != DataType_ULong {
		return 0, errors.New("entry is not a uint32")
	}

	if e.Length != 1 {
		return 0, errors.New("entry length is not 1")
	}

	return e.RawValue, nil
}

// ReadUints32 reads and returns a slice of uint32 from an IFD entry. It returns an error if it cannot read the slice.
func (e Entry) ReadUints32(s Source) ([]uint32, error) {
	if e.DataType != DataType_ULong {
		return nil, errors.New("entry is not a []uint32")
	}

	if e.Length == 1 {
		return []uint32{e.RawValue}, nil
	}

	res := make([]uint32, e.Length)
	if err := s.SeekFromStart(int64(e.RawValue)); err != nil {
		return nil, err
	}

	size := 4
	buffer := make([]byte, size*int(e.Length))
	if err := s.ReadInto(buffer); err != nil {
		return nil, err
	}

	for i := 0; i < int(e.Length); i++ {
		res[i] = s.Uint32(buffer[i*size : i*size+size])
	}

	return res, nil
}

// ReadURational reads and returns an unsigned rational from an IFD entry, returning its numerator and denominator as uint32. It returns an error if it cannot read from the underlying reader.
func (e Entry) ReadURational(s Source) (uint32, uint32, error) {
	if e.DataType != DataType_URational {
		return 0, 0, errors.New("entry is not a rational")
	}

	if err := s.SeekFromStart(int64(e.RawValue)); err != nil {
		return 0, 0, err
	}

	buffer := make([]byte, 8)
	if err := s.ReadInto(buffer); err != nil {
		return 0, 0, err
	}

	numerator := s.Uint32(buffer[0:4])
	denominator := s.Uint32(buffer[4:8])

	return numerator, denominator, nil
}

func (e Entry) PrintValue(s Source) (string, error) {
	switch e.DataType {
	case DataType_UByte:
		return "cannot yet print ubyte", nil
	case DataType_String:
		return e.ReadString(s)
	case DataType_UShort:
		values, err := e.ReadUints16(s)
		if err != nil {
			return "", err
		}
		if len(values) == 1 {
			return fmt.Sprintf("%d", values[0]), nil
		}
		return fmt.Sprintf("%v", values), nil
	case DataType_ULong:
		values, err := e.ReadUints32(s)
		if err != nil {
			return "", err
		}
		if len(values) == 1 {
			return fmt.Sprintf("%d", values[0]), nil
		}
		return fmt.Sprintf("%v", values), nil
	case DataType_URational:
		num, den, err := e.ReadURational(s)
		if err != nil {
			return "", err
		}
		if den == 1 {
			return fmt.Sprintf("%d", num), nil
		}
		return fmt.Sprintf("%d / %d", num, den), nil
	case DataType_Byte:
		return "cannot yet print byte", nil
	case DataType_UByte_Sequence:
		return "cannot yet print ubyte sequence", nil
	case DataType_Short:
		return "cannot yet print int16", nil
	case DataType_Long:
		return "cannot yet print int32", nil
	case DataType_Rational:
		return "cannot yet print rational", nil
	case DataType_Single_Precision_IEEE_Format:
		return "cannot yet print single-precision float", nil
	case DataType_Double_Precision_IEEE_Format:
		return "cannot yet print double-precision float", nil
	default:
		return "", nil
	}
}
