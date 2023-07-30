package entry

type ID uint16

// Entry represents an IFD entry
type Entry struct {
	ID       ID
	DataType uint16
	Length   uint32
	Value    uint32 // value of the entry or byte offset to read the value from, depending on the DataType and Length
}

const (
	// IFD #0

	ImageWidth                ID = 0x100
	ImageHeight               ID = 0x101
	BitsPerSample             ID = 0x102
	Compression               ID = 0x103
	PhotometricInterpretation ID = 0x106
	Thresholding              ID = 0x107
	CellWidth                 ID = 0x108
	CellLength                ID = 0x109
	FillOrder                 ID = 0x10a
	DocumentName              ID = 0x10d
	ImageDescription          ID = 0x10e
	Make                      ID = 0x10f
	Model                     ID = 0x110
	Exif                      ID = 0x8769
	GPSInfo                   ID = 0x8825

	// Exif sub-IFD

	DateTimeOriginal   ID = 0x9003
	OffsetTimeOriginal ID = 0x9011

	// GPSInfo sub-IFD

	GPSLatitudeRef = 0x0001
)