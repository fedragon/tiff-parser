package tiff

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

	// ExifOffsetId is the offset ID of the EXIF Sub-IFD
	ExifOffsetId = 0x8769

	// dateTimeOriginal is the Datetime when original photo was taken
	dateTimeOriginal = 0x9003
	// offsetTimeOriginal is the name of dateTimeOriginal's timezone (e.g. Europe/Amsterdam)
	offsetTimeOriginal = 0x9011
)
