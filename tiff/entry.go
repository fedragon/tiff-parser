package tiff

import (
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
	MakerNotes         EntryID = 0x927c

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
)

type URational struct {
	Numerator   uint32
	Denominator uint32
}

type Rational struct {
	Numerator   int32
	Denominator int32
}

type EntryValue struct {
	UByte     *byte
	String    *string
	Uint16    *uint16
	Uints16   []uint16
	Uint32    *uint32
	Uints32   []uint32
	URational *URational
	Byte      *byte
	Int16     *int16
	Ints16    []int16
	Int32     *int32
	Ints32    []int32
	Rational  *Rational
}

// Entry represents an IFD entry
type Entry struct {
	ID       EntryID
	DataType DataType
	Length   uint32
	RawValue uint32 // value of the entry or offset to read the value from, depending on DataType and Length
	Value    EntryValue
}

func (e Entry) String() string {
	dt := "UNKNOWN"
	value := "not yet implemented"
	switch DataType(e.DataType) {
	case DataType_UByte:
		dt = "unsigned byte"
		value = fmt.Sprintf("%d", *e.Value.UByte)
	case DataType_String:
		dt = "string"
		value = *e.Value.String
	case DataType_UShort:
		dt = "unsigned short 16bits"
		if e.Length == 1 {
			value = fmt.Sprintf("%d", *e.Value.Uint16)
		} else {
			value = fmt.Sprintf("%v", e.Value.Uints16)
		}
	case DataType_ULong:
		dt = "unsigned long 32bits"
		if e.Length == 1 {
			value = fmt.Sprintf("%d", *e.Value.Uint32)
		} else {
			value = fmt.Sprintf("%v", e.Value.Uints32)
		}
	case DataType_URational:
		dt = "unsigned rational"
		value = fmt.Sprintf("%d / %d", e.Value.URational.Numerator, e.Value.URational.Denominator)
	case DataType_Byte:
		dt = "signed byte"
		value = fmt.Sprintf("%d", *e.Value.Byte)
	case DataType_UByte_Sequence:
		dt = "unsigned byte sequence"
	case DataType_Short:
		dt = "signed short 16bits"
		if e.Length == 1 {
			value = fmt.Sprintf("%d", *e.Value.Int16)
		} else {
			value = fmt.Sprintf("%v", e.Value.Ints16)
		}
	case DataType_Long:
		dt = "signed long 32bits"
		if e.Length == 1 {
			value = fmt.Sprintf("%d", *e.Value.Int32)
		} else {
			value = fmt.Sprintf("%v", e.Value.Ints32)
		}
	case DataType_Rational:
		dt = "signed rational"
		value = fmt.Sprintf("%d / %d", e.Value.Rational.Numerator, e.Value.Rational.Denominator)
	}

	return fmt.Sprintf("ID: 0x%X\nDataType: %s\nLength: %d\nValue: %s\n", e.ID, dt, e.Length, value)
}
