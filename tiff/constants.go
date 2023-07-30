package tiff

type EntryID uint16

const (
	// IntelByteOrder is the TIFF standard value to indicate Intel byte ordering (aka little-endian)
	IntelByteOrder = 0x4949
	// MotorolaByteOrder is the TIFF standard value to indicate Motorola byte ordering (aka big-endian)
	MotorolaByteOrder = 0x4D4D

	// MagicNumberBigEndian is the TIFF standard value to indicate big-endian byte ordering
	MagicNumberBigEndian = 0x002A
	// MagicNumberLittleEndian is the TIFF standard value to indicate little-endian byte ordering
	MagicNumberLittleEndian = 0x2A00

	// OrfMagicNumberBigEndian is the ORF-specific value to indicate big-endian byte ordering
	OrfMagicNumberBigEndian = 0x4F52
	// OrfMagicNumberLittleEndian is the ORF-specific value to indicate little-endian byte ordering
	OrfMagicNumberLittleEndian = 0x524F

	ExifOffset         EntryID = 0x8769
	DateTimeOriginal   EntryID = 0x9003
	OffsetTimeOriginal EntryID = 0x9011
)
