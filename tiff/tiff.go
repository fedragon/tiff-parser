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

// ParseOriginalDatetime parses the original date/time from the EXIF data of a TIFF file. Several vendor-specific formats conform to the TIFF header structure (e.g. CR2, ORF).
func ParseOriginalDatetime(r io.ReadSeeker) (time.Time, bool, error) {
	header := make([]byte, 8) // only read the first 3 fields of the TIFF header
	_, err := io.ReadFull(r, header)
	if err != nil {
		return time.Time{}, false, err
	}

	byteOrder, err := readEndianness(header[0:2])
	if err != nil {
		return time.Time{}, false, err
	}

	if err := validateMagicNumber(byteOrder, header[2:4]); err != nil {
		return time.Time{}, false, err
	}

	offsetToFirstIfd := int64(byteOrder.Uint32(header[4:8]))
	if _, err := r.Seek(offsetToFirstIfd, io.SeekStart); err != nil {
		return time.Time{}, false, err
	}

	dt, err := readOriginalDateTime(byteOrder, r, offsetToFirstIfd)
	if err != nil {
		return time.Time{}, false, err
	}

	return dt, true, nil
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

// readOriginalDateTime reads and returns the original date/time from the EXIF subdirectory of IFD#0. It also reads and uses the timezone offset, if available.
func readOriginalDateTime(byteOrder binary.ByteOrder, r io.ReadSeeker, offset int64) (time.Time, error) {
	entries, err := collectIFDEntries(byteOrder, r, offset, newWanted(ExifOffset))
	if err != nil {
		return time.Time{}, err
	}

	exifOffset, ok := entries[ExifOffset]
	if !ok {
		return time.Time{}, errors.New("no exif data")
	}

	entries, err = collectIFDEntries(byteOrder, r, int64(exifOffset.Value), newWanted(DateTimeOriginal, OffsetTimeOriginal))
	if err != nil {
		return time.Time{}, err
	}

	dateTimeOriginal, ok := entries[DateTimeOriginal]
	if !ok {
		return time.Time{}, errors.New("DateTimeOriginal not found")
	}

	dateTimeString, err := readString(dateTimeOriginal, r)
	if err != nil {
		return time.Time{}, err
	}

	dateTime, err := time.Parse("2006:01:02 15:04:05", dateTimeString)
	if err != nil {
		return time.Time{}, err
	}

	offsetTime, ok := entries[OffsetTimeOriginal]
	if !ok {
		return dateTime, nil
	}

	offsetTimeString, err := readString(offsetTime, r)
	if err != nil {
		return time.Time{}, err
	}

	loc, err := time.LoadLocation(offsetTimeString)
	if err != nil {
		return time.Time{}, err
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
	), nil
}

// collectIFDEntries collects a set of IFD entries from an IFD.
// To save memory and time (an IFD may contain tens of thousands of entries), it returns as soon as:
// - all entries have been collected, or
// - it has scanned the maximum ID among the desired ones (entries are written according to the natural ordering of their
// ID value: no point in looking further).
func collectIFDEntries(byteOrder binary.ByteOrder, r io.ReadSeeker, offset int64, wanted *wanted) (map[EntryID]Entry, error) {
	var entries = make(map[EntryID]Entry)

	buffer := make([]byte, 2)
	_, err := r.Read(buffer)
	if err != nil {
		return nil, err
	}
	numEntries := int64(byteOrder.Uint16(buffer))

	for i := int64(0); i < numEntries; i++ {
		buffer := make([]byte, 12)
		if _, err := r.Seek(offset+2+i*12, io.SeekStart); err != nil {
			return nil, err
		}
		_, err := io.ReadFull(r, buffer)
		if err != nil {
			return nil, err
		}

		id := EntryID(byteOrder.Uint16(buffer[:2]))
		if wanted.Contains(id) {
			entries[id] = Entry{
				ID:       id,
				DataType: byteOrder.Uint16(buffer[2:4]),
				Length:   byteOrder.Uint32(buffer[4:8]),
				Value:    byteOrder.Uint32(buffer[8:12]),
			}
		}

		// No point in scanning the IFD further: if we've already found all desired IDs, we're done; if not, we're not going to find them further anyway
		if id >= wanted.Max() {
			break
		}
	}

	return entries, nil
}

// readString reads and returns a string from an IFD entry, trimming its NUL-byte terminator.
func readString(entry Entry, r io.ReadSeeker) (string, error) {
	if _, err := r.Seek(int64(entry.Value), io.SeekStart); err != nil {
		return "", err
	}
	res := make([]byte, entry.Length)
	if _, err := r.Read(res); err != nil {
		return "", err
	}

	return string(bytes.TrimSuffix(res, []byte{0x0})), nil
}
