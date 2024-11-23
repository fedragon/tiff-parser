package main

import (
	"fmt"
	"os"

	"github.com/fedragon/tiff-parser/tiff"
)

func main() {
	r, err := os.Open("<file_path>")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	p, err := tiff.NewParser(r)
	if err != nil {
		panic(err)
	}

	// provide the ID(s) of the entry you would like to collect (multiple IDs allowed)
	entries, err := p.Parse(tiff.ImageWidth)
	if err != nil {
		panic(err)
	}

	if en, ok := entries[tiff.ImageWidth]; ok {
		// if you're sure about the type of this field
		fmt.Println("width", *en.Value.Uint16)

		// otherwise
		switch en.DataType {
		case tiff.DataType_UShort:
			fmt.Println("width", *en.Value.Uint16)
			// other cases ...
		}
	}

	// an `EntryID` is simply a type alias for `uint16`
	model := tiff.EntryID(0x0110)

	// only needed when your entry is not listed in `tiff.Defaults`
	p.WithMapping(map[tiff.EntryID]tiff.Group{
		model: tiff.Group_IFD0, // Model belongs to the first IFD (aka IFD#0)
		// ...
	})

	entries, err = p.Parse(model)
	if err != nil {
		panic(err)
	}

	fmt.Println("model", entries[model].Value.String)
}
