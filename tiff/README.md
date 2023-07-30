## TIFF parser

Parses Exif metadata from TIFF-like files. Tested on Canon's CR2 and Olympus' ORF files.

### Usage

```go
package main

import (
    "fmt"
    "os"

    "github.com/fedragon/tiff-parser/tiff"
    "github.com/fedragon/tiff-parser/tiff/entry"
)

func main() {
	r, err := os.Open("...")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	p, err := tiff.NewParser(r)
	if err != nil {
		panic(err)
	}

	// required to know in which (sub-)IFD to find entries: if your entry is not listed in `tiff.Defaults`, add it here
	p.WithMapping(map[entry.ID]tiff.Group {
		entry.Model: tiff.GroupIfd0,
		// ...
    })

	// provide the IDs of the entries you would like to collect
	entries, err := p.Parse(entry.ImageWidth, entry.ImageHeight, entry.Make)
	if err != nil {
		panic(err)
	}

	// read the value, casting it to the expected data type (look at References for a link to ExifTool's compendium of all entries)
	width, err := p.ReadUint32(entries[entry.ImageWidth])
	if err != nil {
		panic(err)
	}
	fmt.Printf("width: %v\n", width)

	maker, err := p.ReadString(entries[entry.Make])
	if err != nil {
		panic(err)
	}
	fmt.Printf("make: %v\n", maker)
}
```

### TIFF File structure

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

### References

[ExifTool - Exif Tags](https://exiftool.org/TagNames/EXIF.html)