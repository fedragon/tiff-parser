# TIFF parser

Parses Exif metadata from TIFF-like files. Tested on Canon's CR2 and Olympus' ORF files.

## Usage

This library provides a low-level API to parse IFD entries from TIFF files: low-level here means that clients of this library need to provide the _IFD Entry ID_ of any entry they would like to retrieve and they also need to know what is its datatype, so that they can later read it into the appropriate type using the provided `parser.Read*` methods.

This is because there are many manufacturer-specific exceptions to how IFD entries are written, even for basic entries such as `imageWidth` (`uint16` in CR2, `uint32` in ORF).

[ExifTool - Exif Tags](https://exiftool.org/TagNames/EXIF.html) provides a very extensive compendium of all known IFD entries.

### Example

See [examples/main.go](examples/main.go)

## TIFF File structure

```
TIFF header: 8 bytes
CR2 header:  8 bytes
IFD#0:
    Number of entries: 2 bytes, ushort
    Entry#1: 12 bytes per entry
    ...
    Entry_Exif (pointer to the Exif sub-IFD)
    Entry_GPSInfo (pointer to the GPS Info sub-IFD)
    ...
    Entry#N
    Pointer to IFD#1: 4 bytes, ulong
    Exif sub-IFD:
        Number of entries: 2 bytes, ushort
        Entry#1: 12 bytes per entry
        ...
        Entry_MakerNote (pointer to the MakerNote sub-IFD)
        ...
        Entry#N
        MakerNote sub-IFD:
            sequence of bytes containing manufacturer-specific entries, structure depends on manufacturer
    GPSInfo sub-IFD: (it may be before or after Exif, it all depends on their respective pointers' values)
        Number of entries: 2 bytes
        Entry#1 ... Entry#N: 12 bytes per entry
    IFD#0 Data Area: variable size
IFD#1:
    ...
IFD#2:
    ...
IFD#3:
    ...
Image#0:
    ...
Image#1:
    ...
Image#2:
    ...
Image#3:
    ...
```

## Credits

The biggest challenge in parsing Exif metadata is that its structure is poorly documented and there are tons of manufacturer-specific exceptions to take into account.

I was able to write this library only thanks to the excellent work created, over the years, by people like Phil Harvey ([ExifTool](https://exiftool.org)) and Laurent Clévy ([Understanding what is stored in a Canon RAW .CR2 file, how and why](http://lclevy.free.fr/cr2/)).
