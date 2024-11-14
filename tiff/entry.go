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

func (e Entry) String() string {
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

	return fmt.Sprintf("ID: 0x%X\nDataType: %s\nLength: %d\n", e.ID, dt, e.Length)
}
