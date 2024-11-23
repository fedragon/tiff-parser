package tiff

type (
	// Group enumerates known (sub-)IFD: an Image File Directory (IFD) is a physical group of entries.
	Group uint8
)

const (
	Group_IFD0 Group = iota
	Group_IFD1
	Group_Exif
	Group_GPSInfo

	// intelByteOrder is the TIFF standard value to indicate Intel byte ordering (aka little-endian)
	intelByteOrder = 0x4949
	// motorolaByteOrder is the TIFF standard value to indicate Motorola byte ordering (aka big-endian)
	motorolaByteOrder = 0x4D4D

	// magicNumberBigEndian is the TIFF standard value to indicate big-endian byte ordering
	magicNumberBigEndian = 0x002A
	// magicNumberLittleEndian is the TIFF standard value to indicate little-endian byte ordering
	magicNumberLittleEndian = 0x2A00

	// orfMagicNumberBigEndian is the ORF-specific value to indicate big-endian byte ordering
	orfMagicNumberBigEndian = 0x4F52
	// orfMagicNumberLittleEndian is the ORF-specific value to indicate little-endian byte ordering
	orfMagicNumberLittleEndian = 0x524F
)

// Defaults maps IFD entries to the Group they belong to (e.g. IFD#0, Exif, GPSInfo), so that a `Parser` will know where to look for them.
var Defaults = map[EntryID]Group{
	ImageWidth:         Group_IFD0,
	ImageHeight:        Group_IFD0,
	BitsPerSample:      Group_IFD0,
	Compression:        Group_IFD0,
	Make:               Group_IFD0,
	Model:              Group_IFD0,
	Exif:               Group_IFD0,
	GPSInfo:            Group_IFD0,
	ExposureTime:       Group_Exif,
	FNumber:            Group_Exif,
	ISO:                Group_Exif,
	DateTimeOriginal:   Group_Exif,
	OffsetTimeOriginal: Group_Exif,
	GPSLatitude:        Group_GPSInfo,
	GPSLongitude:       Group_GPSInfo,
}
