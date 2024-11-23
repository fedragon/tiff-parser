package tiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Parser represents a TIFF parser
type Parser struct {
	reader         io.ReadSeeker
	byteOrder      binary.ByteOrder
	firstIFDOffset int64
	mapping        map[EntryID]Group
}

// NewParser returns a new parser or an error if the content is not a valid TIFF.
func NewParser(r io.ReadSeeker) (*Parser, error) {
	header := make([]byte, 8) // only read the TIFF header
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	byteOrder, err := readEndianness(header[0:2])
	if err != nil {
		return nil, err
	}

	if err := validateMagicNumber(byteOrder, header[2:4]); err != nil {
		return nil, err
	}

	return &Parser{
		reader:         r,
		byteOrder:      byteOrder,
		firstIFDOffset: int64(byteOrder.Uint32(header[4:8])),
		mapping:        Defaults,
	}, nil
}

// WithMapping adds entry mapping(s) to the parser, so that it will know where those entries appear in the file.
func (p *Parser) WithMapping(m map[EntryID]Group) *Parser {
	for k, v := range m {
		p.mapping[k] = v
	}

	return p
}

// Parse parses the TIFF file, returning any entry found in it that matches the given IDs or an error if the read fails. It does not return an error if one or more of the entries are not found.
func (p *Parser) Parse(ids ...EntryID) (map[EntryID]Entry, error) {
	entries := make(map[EntryID]Entry)
	ifd0Wanted := newWanted()
	exifWanted := newWanted()
	gpsInfoWanted := newWanted()

	for _, id := range ids {
		group, ok := p.mapping[id]

		if ok {
			switch group {
			case Group_IFD0:
				ifd0Wanted.Put(id)
			case Group_Exif:
				ifd0Wanted.Put(Exif)
				exifWanted.Put(id)
			case Group_GPSInfo:
				ifd0Wanted.Put(GPSInfo)
				gpsInfoWanted.Put(id)
			}
		}
	}

	ifd0Offset := p.firstIFDOffset
	ifd0Entries, err := p.collect(ifd0Offset, ifd0Wanted)
	if err != nil {
		return nil, err
	}

	for key, value := range ifd0Entries {
		if key != Exif && key != GPSInfo {
			entries[key] = value
		}
	}

	if !exifWanted.Empty() {
		exifEntry, ok := ifd0Entries[Exif]
		if !ok {
			return nil, errors.New("exif IFD not found")
		}

		exifEntries, err := p.collect(int64(exifEntry.RawValue), exifWanted)
		if err != nil {
			return nil, err
		}

		for key, value := range exifEntries {
			entries[key] = value
		}
	}

	if !gpsInfoWanted.Empty() {
		gpsInfoEntry, ok := ifd0Entries[GPSInfo]
		if !ok {
			return nil, errors.New("exif IFD not found")
		}

		gpsInfoEntries, err := p.collect(int64(gpsInfoEntry.RawValue), gpsInfoWanted)
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
	case intelByteOrder:
		return binary.LittleEndian, nil
	case motorolaByteOrder:
		return binary.BigEndian, nil
	default:
		return nil, fmt.Errorf("unknown endianness: 0x%X", value)
	}
}

// validateMagicNumber validates the file type by checking that it conforms to one of the expected values
func validateMagicNumber(byteOrder binary.ByteOrder, buffer []byte) error {
	magicNumber := byteOrder.Uint16(buffer)
	if magicNumber != magicNumberBigEndian &&
		magicNumber != magicNumberLittleEndian &&
		magicNumber != orfMagicNumberBigEndian &&
		magicNumber != orfMagicNumberLittleEndian {
		return fmt.Errorf("unknown magic number: 0x%X", magicNumber)
	}
	return nil
}

// collect collects a set of IFD entries from an IFD.
// To save memory and time (an IFD may contain tens of thousands of entries), it returns as soon as:
// - all entries have been collected, or
// - it has scanned the maximum ID among the desired ones (entries are written according to the natural ordering of their
// ID value: no point in looking further).
func (p *Parser) collect(startingOffset int64, wanted *wanted) (map[EntryID]Entry, error) {
	offset := startingOffset
	if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	var entries = make(map[EntryID]Entry)
	buffer := make([]byte, 2)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return nil, err
	}
	numEntries := int64(p.byteOrder.Uint16(buffer))
	offset += 2

	for i := int64(0); i < numEntries; i++ {
		buffer := make([]byte, EntryLength)
		if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(p.reader, buffer); err != nil {
			return nil, err
		}
		offset += EntryLength

		id := EntryID(p.byteOrder.Uint16(buffer[:2]))
		if wanted.Contains(id) {
			dt := DataType(p.byteOrder.Uint16(buffer[2:4]))
			length := p.byteOrder.Uint32(buffer[4:8])
			rawValue := p.byteOrder.Uint32(buffer[8:12])
			value, err := p.readValue(dt, length, rawValue)
			if err != nil {
				return nil, err
			}

			entries[id] = Entry{
				ID:       id,
				DataType: dt,
				Length:   length,
				RawValue: rawValue,
				Value:    value,
			}
		}

		if id >= wanted.Max() {
			break
		}
	}

	return entries, nil
}

func (p *Parser) readValue(dt DataType, length uint32, rawValue uint32) (EntryValue, error) {
	switch dt {
	case DataType_UByte:
		value := byte(rawValue)
		return EntryValue{UByte: &value}, nil
	case DataType_String:
		value, err := p.readString(length, rawValue)
		if err != nil {
			return EntryValue{}, err
		}
		return EntryValue{String: &value}, nil
	case DataType_UShort:
		if length == 1 {
			value := uint16(rawValue)
			return EntryValue{Uint16: &value}, nil
		} else {
			values, err := p.readUints16(length, rawValue)
			if err != nil {
				return EntryValue{}, err
			}
			return EntryValue{Uints16: values}, nil
		}
	case DataType_ULong:
		if length == 1 {
			value := rawValue
			return EntryValue{Uint32: &value}, nil
		} else {
			values, err := p.readUints32(length, rawValue)
			if err != nil {
				return EntryValue{}, err
			}
			return EntryValue{Uints32: values}, nil
		}
	case DataType_URational:
		value, err := p.readURational(rawValue)
		if err != nil {
			return EntryValue{}, err
		}
		return EntryValue{URational: &value}, nil
	case DataType_Byte:
		value := byte(rawValue)
		return EntryValue{Byte: &value}, nil
	case DataType_Short:
		if length == 1 {
			value := int16(rawValue)
			return EntryValue{Int16: &value}, nil
		} else {
			values, err := p.readInts16(length, rawValue)
			if err != nil {
				return EntryValue{}, err
			}
			return EntryValue{Ints16: values}, nil
		}
	case DataType_Long:
		if length == 1 {
			value := int32(rawValue)
			return EntryValue{Int32: &value}, nil
		} else {
			values, err := p.readInts32(length, rawValue)
			if err != nil {
				return EntryValue{}, err
			}
			return EntryValue{Ints32: values}, nil
		}
	case DataType_Rational:
		value, err := p.readRational(rawValue)
		if err != nil {
			return EntryValue{}, err
		}
		return EntryValue{Rational: &value}, nil
	}
	return EntryValue{}, nil
}

// readString reads and returns a string from an IFD entry, trimming its NUL-byte terminator. It returns an error if it cannot read the string.
func (p *Parser) readString(length uint32, offset uint32) (string, error) {
	if _, err := p.reader.Seek(int64(offset), io.SeekStart); err != nil {
		return "", err
	}

	buffer := make([]byte, length)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return "", err
	}

	return string(bytes.TrimSuffix(buffer, []byte{0x0})), nil
}

// readUints16 reads and returns a slice of uint16 from an IFD entry. It returns an error if it cannot read the slice.
func (p *Parser) readUints16(length uint32, offset uint32) ([]uint16, error) {
	res := make([]uint16, length)
	if _, err := p.reader.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, err
	}

	size := 2
	buffer := make([]byte, size*int(length))
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return nil, err
	}

	for i := 0; i < int(length); i++ {
		res[i] = p.byteOrder.Uint16(buffer[i*size : i*size+size])
	}

	return res, nil
}

// readInts16 reads and returns a slice of int16 from an IFD entry. It returns an error if it cannot read the slice.
func (p *Parser) readInts16(length uint32, offset uint32) ([]int16, error) {
	uints16, err := p.readUints16(length, offset)
	if err != nil {
		return nil, err
	}
	ints16 := make([]int16, len(uints16))
	for i, u := range uints16 {
		ints16[i] = int16(u)
	}

	return ints16, nil
}

// readUints32 reads and returns a slice of uint32 from an IFD entry. It returns an error if it cannot read the slice.
func (p *Parser) readUints32(length uint32, offset uint32) ([]uint32, error) {
	res := make([]uint32, length)
	if _, err := p.reader.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, err
	}

	size := 4
	buffer := make([]byte, size*int(length))
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return nil, err
	}

	for i := 0; i < int(length); i++ {
		res[i] = p.byteOrder.Uint32(buffer[i*size : i*size+size])
	}

	return res, nil
}

// readInts32 reads and returns a slice of int32 from an IFD entry. It returns an error if it cannot read the slice.
func (p *Parser) readInts32(length uint32, offset uint32) ([]int32, error) {
	uints32, err := p.readUints32(length, offset)
	if err != nil {
		return nil, err
	}
	ints32 := make([]int32, len(uints32))
	for i, u := range uints32 {
		ints32[i] = int32(u)
	}

	return ints32, nil
}

// readURational reads and returns an unsigned rational from an IFD entry, representing its numerator and denominator as uint32. It returns an error if it cannot read from the underlying reader.
func (p *Parser) readURational(offset uint32) (URational, error) {
	if _, err := p.reader.Seek(int64(offset), io.SeekStart); err != nil {
		return URational{}, err
	}

	buffer := make([]byte, 8)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return URational{}, err
	}

	return URational{p.byteOrder.Uint32(buffer[0:4]), p.byteOrder.Uint32(buffer[4:8])}, nil
}

// readRational reads and returns an signed rational from an IFD entry, representing its numerator and denominator as int32. It returns an error if it cannot read from the underlying reader.
func (p *Parser) readRational(offset uint32) (Rational, error) {
	if _, err := p.reader.Seek(int64(offset), io.SeekStart); err != nil {
		return Rational{}, err
	}

	buffer := make([]byte, 8)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return Rational{}, err
	}

	return Rational{int32(p.byteOrder.Uint32(buffer[0:4])), int32(p.byteOrder.Uint32(buffer[4:8]))}, nil
}

func printEntries(p *Parser, offsets []int64) error {
	if len(offsets) == 0 {
		return nil
	}

	offset := offsets[0]
	offsets = offsets[1:]
	if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	buffer := make([]byte, 2)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return err
	}
	numEntries := p.byteOrder.Uint16(buffer)
	offset += 2

	for i := uint16(0); i < numEntries; i++ {
		buffer := make([]byte, EntryLength)
		if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
			return err
		}
		if _, err := io.ReadFull(p.reader, buffer); err != nil {
			return err
		}

		id := EntryID(p.byteOrder.Uint16(buffer[:2]))
		dt := DataType(p.byteOrder.Uint16(buffer[2:4]))
		length := p.byteOrder.Uint32(buffer[4:8])
		rawValue := p.byteOrder.Uint32(buffer[8:12])
		value, err := p.readValue(dt, length, rawValue)
		if err != nil {
			return err
		}

		entry := Entry{
			ID:       id,
			DataType: dt,
			Length:   length,
			RawValue: rawValue,
			Value:    value,
		}

		if entry.ID == Exif {
			fmt.Println("exif offset", entry.RawValue)
			offsets = append(offsets, int64(entry.RawValue))
		} else if entry.ID == GPSInfo {
			fmt.Println("gps offset", entry.RawValue)
			offsets = append(offsets, int64(entry.RawValue))
		}

		fmt.Println(entry.String())
		offset += EntryLength
	}

	if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	buffer = make([]byte, 4)
	n, err := io.ReadFull(p.reader, buffer)
	if err != nil {
		return err
	}
	if n == 4 {
		next := p.byteOrder.Uint32(buffer)
		if next > 0 {
			fmt.Println("appending offset", next)
			offsets = append(offsets, int64(next))
		}
	}

	return printEntries(p, offsets)
}

func (p *Parser) PrintEntries() error {
	return printEntries(p, []int64{p.firstIFDOffset})
}

// ReadThumbnail reads the thumbnail stored in Image Data #1. The offset and length of Image Data #1 are written in IFD #1.
func (p *Parser) ReadThumbnail() ([]byte, error) {
	offset := p.firstIFDOffset
	if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	buffer := make([]byte, 2)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return nil, err
	}

	// skip all entries in this IFD
	numEntries := int64(p.byteOrder.Uint16(buffer))
	offset += 2 + numEntries*12
	if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	buffer = make([]byte, 4) // offset to IFD#1 is a ulong
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return nil, err
	}
	offsetToIdf1 := int64(p.byteOrder.Uint32(buffer))

	entries, err := p.collect(
		offsetToIdf1,
		newWanted(ThumbnailOffset, ThumbnailLength),
	)
	if err != nil {
		return nil, err
	}

	var thumbnailOffset uint32
	if elem, ok := entries[ThumbnailOffset]; ok {
		if offset := elem.Value.Uint32; offset != nil {
			thumbnailOffset = *offset
		} else {
			return nil, errors.New("thumbnail offset not found")
		}
	}

	var thumbnailLength uint32
	if elem, ok := entries[ThumbnailLength]; ok {
		if length := elem.Value.Uint32; length != nil {
			thumbnailLength = *length
		} else {
			return nil, errors.New("thumbnail length not found")
		}
	}

	if _, err := p.reader.Seek(int64(thumbnailOffset), io.SeekStart); err != nil {
		return nil, err
	}
	buffer = make([]byte, thumbnailLength)
	if _, err := io.ReadFull(p.reader, buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}
