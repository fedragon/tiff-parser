package tiff

import "github.com/fedragon/tiff-parser/tiff/entry"

type (
	// Group enumerates known (sub-)IFD: an Image File Directory (IFD) is a physical group of entries.
	Group uint8
)

const (
	GroupIfd0 Group = iota
	GroupExif
	GroupGPSInfo

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
)

// Defaults maps IFD entries to the Group they belong to (e.g. IFD#0, Exif, GPSInfo), so that a `Parser` will know where to look for them.
var Defaults = map[entry.ID]Group{
	entry.ImageWidth:         GroupIfd0,
	entry.ImageHeight:        GroupIfd0,
	entry.BitsPerSample:      GroupIfd0,
	entry.Compression:        GroupIfd0,
	entry.Make:               GroupIfd0,
	entry.Model:              GroupIfd0,
	entry.Exif:               GroupIfd0,
	entry.GPSInfo:            GroupIfd0,
	entry.ExposureTime:       GroupExif,
	entry.FNumber:            GroupExif,
	entry.ISO:                GroupExif,
	entry.DateTimeOriginal:   GroupExif,
	entry.OffsetTimeOriginal: GroupExif,
	entry.GPSLatitude:        GroupGPSInfo,
	entry.GPSLongitude:       GroupGPSInfo,
}
