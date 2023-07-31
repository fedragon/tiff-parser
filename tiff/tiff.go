package tiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/fedragon/tiff-parser/tiff/entry"
)

// Parser represents a TIFF parser
type Parser struct {
	reader         io.ReadSeeker
	byteOrder      binary.ByteOrder
	firstIFDOffset int64
	mapping        map[entry.ID]Group
}

// NewParser returns a new parser or an error if the content is not a valid TIFF.
func NewParser(r io.ReadSeeker) (*Parser, error) {
	header := make([]byte, 8) // only read the TIFF header
	_, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}

	byteOrder, err := readEndianness(header[0:2])
	if err != nil {
		return nil, err
	}

	if err := validateMagicNumber(byteOrder, header[2:4]); err != nil {
		return nil, err
	}

	firstIDOffset := int64(byteOrder.Uint32(header[4:8]))

	return &Parser{
		reader:         r,
		byteOrder:      byteOrder,
		firstIFDOffset: firstIDOffset,
		mapping:        Defaults,
	}, nil
}

// WithMapping adds entry mapping(s) to the parser
func (p *Parser) WithMapping(m map[entry.ID]Group) *Parser {
	for k, v := range m {
		p.mapping[k] = v
	}

	return p
}

// Parse parses the TIFF file, returning any entry found in it or an error if the read fails.
func (p *Parser) Parse(ids ...entry.ID) (map[entry.ID]entry.Entry, error) {
	entries := make(map[entry.ID]entry.Entry)
	ifd0Wanted := newWanted()
	exifWanted := newWanted()
	gpsInfoWanted := newWanted()

	for _, id := range ids {
		group, ok := p.mapping[id]

		if ok {
			switch group {
			case GroupIfd0:
				ifd0Wanted.Put(id)
			case GroupExif:
				ifd0Wanted.Put(entry.Exif)
				exifWanted.Put(id)
			case GroupGPSInfo:
				ifd0Wanted.Put(entry.GPSInfo)
				gpsInfoWanted.Put(id)
			}
		}
	}

	ifd0Entries, err := p.collect(p.firstIFDOffset, ifd0Wanted)
	if err != nil {
		return nil, err
	}

	for key, value := range ifd0Entries {
		if key != entry.Exif && key != entry.GPSInfo {
			entries[key] = value
		}
	}

	if !exifWanted.Empty() {
		exifEntry, ok := ifd0Entries[entry.Exif]
		if !ok {
			return nil, errors.New("exif IFD not found")
		}

		exifEntries, err := p.collect(int64(exifEntry.Value), exifWanted)
		if err != nil {
			return nil, err
		}

		for key, value := range exifEntries {
			entries[key] = value
		}
	}

	if !gpsInfoWanted.Empty() {
		gpsInfoEntry, ok := ifd0Entries[entry.GPSInfo]
		if !ok {
			return nil, errors.New("exif IFD not found")
		}

		gpsInfoEntries, err := p.collect(int64(gpsInfoEntry.Value), gpsInfoWanted)
		if err != nil {
			return nil, err
		}

		for key, value := range gpsInfoEntries {
			entries[key] = value
		}
	}

	return entries, nil
}

// readEndianness reads and returns the endianness of the metadata.
func readEndianness(buffer []byte) (binary.ByteOrder, error) {
	// Note: the value of these 2 bytes is endianness-independent, so I can use any byte order to read them.
	value := binary.LittleEndian.Uint16(buffer)
	switch value {
	case IntelByteOrder:
		return binary.LittleEndian, nil
	case MotorolaByteOrder:
		return binary.BigEndian, nil
	default:
		return nil, fmt.Errorf("unknown endianness: 0x%X", value)
	}
}

// validateMagicNumber validates the file type by checking that it conforms to one of the expected values
func validateMagicNumber(byteOrder binary.ByteOrder, buffer []byte) error {
	magicNumber := byteOrder.Uint16(buffer)
	if magicNumber != MagicNumberBigEndian &&
		magicNumber != MagicNumberLittleEndian &&
		magicNumber != OrfMagicNumberBigEndian &&
		magicNumber != OrfMagicNumberLittleEndian {
		return fmt.Errorf("unknown magic number: 0x%X", magicNumber)
	}
	return nil
}

// collect collects a set of IFD entries from an IFD.
// To save memory and time (an IFD may contain tens of thousands of entries), it returns as soon as:
// - all entries have been collected, or
// - it has scanned the maximum ID among the desired ones (entries are written according to the natural ordering of their
// ID value: no point in looking further).
func (p *Parser) collect(startingOffset int64, wanted *wanted) (map[entry.ID]entry.Entry, error) {
	offset := startingOffset
	if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	var entries = make(map[entry.ID]entry.Entry)
	buffer := make([]byte, 2)
	_, err := io.ReadFull(p.reader, buffer)
	if err != nil {
		return nil, err
	}
	numEntries := int64(p.byteOrder.Uint16(buffer))
	if _, err := p.reader.Seek(offset+2, io.SeekStart); err != nil {
		return nil, err
	}
	offset += 2

	for i := int64(0); i < numEntries; i++ {
		buffer := make([]byte, entry.Size)
		if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
			return nil, err
		}
		_, err := io.ReadFull(p.reader, buffer)
		if err != nil {
			return nil, err
		}
		offset += entry.Size

		id := entry.ID(p.byteOrder.Uint16(buffer[:2]))
		if wanted.Contains(id) {
			entries[id] = entry.Entry{
				ID:       id,
				DataType: p.byteOrder.Uint16(buffer[2:4]),
				Length:   p.byteOrder.Uint32(buffer[4:8]),
				Value:    p.byteOrder.Uint32(buffer[8:12]),
			}
		}

		if id >= wanted.Max() {
			break
		}
	}

	return entries, nil
}

// ReadString reads and returns a string from an IFD entry, trimming its NUL-byte terminator. It returns an error if it cannot read the string.
func (p *Parser) ReadString(entry entry.Entry) (string, error) {
	if entry.DataType != 2 {
		return "", errors.New("entry is not a string")
	}

	if _, err := p.reader.Seek(int64(entry.Value), io.SeekStart); err != nil {
		return "", err
	}
	res := make([]byte, entry.Length)
	if _, err := io.ReadFull(p.reader, res); err != nil {
		return "", err
	}

	return string(bytes.TrimSuffix(res, []byte{0x0})), nil
}

// ReadUint16 reads and returns a uint16 from an IFD entry. It returns an error if the length is not 1.
func (p *Parser) ReadUint16(entry entry.Entry) (uint16, error) {
	if entry.DataType != 3 {
		return 0, errors.New("entry is not a uint16")
	}

	if entry.Length != 1 {
		return 0, fmt.Errorf("unexpected length: %d", entry.Length)
	}

	return uint16(entry.Value), nil
}

// ReadUints16 reads and returns a slice of uint16 from an IFD entry. It returns an error if it cannot read the slice.
func (p *Parser) ReadUints16(entry entry.Entry) ([]uint16, error) {
	if entry.DataType != 3 {
		return nil, errors.New("entry is not a []uint16")
	}

	if entry.Length == 1 {
		return []uint16{uint16(entry.Value)}, nil
	}

	res := make([]uint16, entry.Length)
	if _, err := p.reader.Seek(int64(entry.Value), io.SeekStart); err != nil {
		return nil, err
	}

	buffer := make([]byte, 2*entry.Length)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return nil, err
	}
	for i := 0; i < int(entry.Length); i++ {
		res[i] = p.byteOrder.Uint16(buffer[i*2 : i*2+2])
	}

	return res, nil
}

// ReadUint32 reads and returns a uint32 from an IFD entry. It returns an error if the length is not 1.
func (p *Parser) ReadUint32(entry entry.Entry) (uint32, error) {
	if entry.DataType != 4 {
		return 0, errors.New("entry is not a uint32")
	}

	if entry.Length != 1 {
		return 0, fmt.Errorf("unexpected length: %d", entry.Length)
	}

	return entry.Value, nil
}

// ReadUints32 reads and returns a slice of uint32 from an IFD entry. It returns an error if it cannot read the slice.
func (p *Parser) ReadUints32(entry entry.Entry) ([]uint32, error) {
	if entry.DataType != 4 {
		return nil, errors.New("entry is not a []uint32")
	}

	if entry.Length == 1 {
		return []uint32{entry.Value}, nil
	}

	res := make([]uint32, entry.Length)
	if _, err := p.reader.Seek(int64(entry.Value), io.SeekStart); err != nil {
		return nil, err
	}

	size := 4

	buffer := make([]byte, size*int(entry.Length))
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return nil, err
	}

	for i := 0; i < int(entry.Length); i++ {
		res[i] = p.byteOrder.Uint32(buffer[i*size : i*size+size])
	}

	return res, nil
}

// ReadURational reads and returns an unsigned rational from an IFD entry, returning its numerator and denominator as uint32. It returns an error if it cannot read from the underlying reader.
func (p *Parser) ReadURational(entry entry.Entry) (uint32, uint32, error) {
	if entry.DataType != 5 {
		return 0, 0, errors.New("entry is not a rational")
	}

	if _, err := p.reader.Seek(int64(entry.Value), io.SeekStart); err != nil {
		return 0, 0, err
	}

	buffer := make([]byte, 8)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return 0, 0, err
	}

	numerator := p.byteOrder.Uint32(buffer[0:4])
	denominator := p.byteOrder.Uint32(buffer[4:8])

	return numerator, denominator, nil
}
