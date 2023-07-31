package entry

type ID uint16

// Entry represents an IFD entry
type Entry struct {
	ID       ID
	DataType uint16
	Length   uint32
	Value    uint32 // value of the entry or offset to read the value from, depending on DataType and Length
}

const (
	// Size of an IFD entry, in bytes
	Size = 12

	// IFD #0

	ImageWidth    ID = 0x100
	ImageHeight   ID = 0x101
	BitsPerSample ID = 0x102
	Compression   ID = 0x103
	Make          ID = 0x10f
	Model         ID = 0x110
	Exif          ID = 0x8769
	GPSInfo       ID = 0x8825

	// Exif sub-IFD

	ExposureTime       ID = 0x829a
	FNumber            ID = 0x829d
	ISO                ID = 0x8827
	DateTimeOriginal   ID = 0x9003
	OffsetTimeOriginal ID = 0x9011

	// GPSInfo sub-IFD

	GPSLatitude  = 0x0002
	GPSLongitude = 0x0004
)
