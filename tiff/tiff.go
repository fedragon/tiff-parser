package tiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Entry represents an IFD entry
type Entry struct {
	ID       EntryID
	DataType uint16
	Length   uint32
	Value    uint32 // value of the entry or byte offset to read the value from, depending on the DataType and Length
}

// Parser represents a TIFF parser
type Parser struct {
	reader    io.ReadSeeker
	byteOrder binary.ByteOrder
	offset    int64
}

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
	if _, err := r.Seek(firstIDOffset, io.SeekStart); err != nil {
		return nil, err
	}

	return &Parser{
		reader:    r,
		byteOrder: byteOrder,
		offset:    firstIDOffset,
	}, nil
}

// ParseOriginalDatetime parses the original date/time from the EXIF data of a TIFF file. Several vendor-specific formats conform to the TIFF header structure (e.g. CR2, ORF).
func (p *Parser) ParseOriginalDatetime() (time.Time, bool, error) {
	entries, err := p.collectIFDEntries(newWanted(ExifOffset))
	if err != nil {
		return time.Time{}, false, err
	}

	exifOffset, ok := entries[ExifOffset]
	if !ok {
		return time.Time{}, false, errors.New("no exif data")
	}
	p.offset = int64(exifOffset.Value)

	entries, err = p.collectIFDEntries(newWanted(DateTimeOriginal, OffsetTimeOriginal))
	if err != nil {
		return time.Time{}, false, err
	}

	dateTimeOriginal, ok := entries[DateTimeOriginal]
	if !ok {
		return time.Time{}, false, errors.New("DateTimeOriginal not found")
	}

	dateTimeString, err := readString(dateTimeOriginal, p.reader)
	if err != nil {
		return time.Time{}, false, err
	}

	dateTime, err := time.Parse("2006:01:02 15:04:05", dateTimeString)
	if err != nil {
		return time.Time{}, false, err
	}

	offsetTime, ok := entries[OffsetTimeOriginal]
	if !ok {
		return dateTime, true, nil
	}

	offsetTimeString, err := readString(offsetTime, p.reader)
	if err != nil {
		return time.Time{}, false, err
	}

	loc, err := time.LoadLocation(offsetTimeString)
	if err != nil {
		return time.Time{}, false, err
	}

	return time.Date(
		dateTime.Year(),
		dateTime.Month(),
		dateTime.Day(),
		dateTime.Hour(),
		dateTime.Minute(),
		dateTime.Second(),
		dateTime.Nanosecond(),
		loc,
	), true, nil
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

// collectIFDEntries collects a set of IFD entries from an IFD.
// To save memory and time (an IFD may contain tens of thousands of entries), it returns as soon as:
// - all entries have been collected, or
// - it has scanned the maximum ID among the desired ones (entries are written according to the natural ordering of their
// ID value: no point in looking further).
func (p *Parser) collectIFDEntries(wanted *wanted) (map[EntryID]Entry, error) {
	var entries = make(map[EntryID]Entry)

	buffer := make([]byte, 2)
	_, err := io.ReadFull(p.reader, buffer)
	if err != nil {
		return nil, err
	}
	numEntries := int64(p.byteOrder.Uint16(buffer))
	if _, err := p.reader.Seek(p.offset+2, io.SeekStart); err != nil {
		return nil, err
	}
	p.offset += 2
	nextIFDOffset := p.offset + numEntries*12

	for i := int64(0); i < numEntries; i++ {
		buffer := make([]byte, 12)
		if _, err := p.reader.Seek(p.offset+12, io.SeekStart); err != nil {
			return nil, err
		}
		p.offset += 12
		_, err := io.ReadFull(p.reader, buffer)
		if err != nil {
			return nil, err
		}

		id := EntryID(p.byteOrder.Uint16(buffer[:2]))
		if wanted.Contains(id) {
			entries[id] = Entry{
				ID:       id,
				DataType: p.byteOrder.Uint16(buffer[2:4]),
				Length:   p.byteOrder.Uint32(buffer[4:8]),
				Value:    p.byteOrder.Uint32(buffer[8:12]),
			}
		}

		// No point in scanning the IFD further: if we've already found all desired IDs, we're done; if not, we're not going to find them further anyway
		if id >= wanted.Max() {
			break
		}
	}

	p.offset = nextIFDOffset

	return entries, nil
}

// readString reads and returns a string from an IFD entry, trimming its NUL-byte terminator.
func readString(entry Entry, r io.ReadSeeker) (string, error) {
	if _, err := r.Seek(int64(entry.Value), io.SeekStart); err != nil {
		return "", err
	}
	res := make([]byte, entry.Length)
	if _, err := io.ReadFull(r, res); err != nil {
		return "", err
	}

	return string(bytes.TrimSuffix(res, []byte{0x0})), nil
}
