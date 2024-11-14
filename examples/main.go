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

	// an `EntryID` is simply a type alias for `uint16`
	model := tiff.EntryID(0x0110)

	// only needed when your entry is not listed in `tiff.Defaults`
	p.WithMapping(map[tiff.EntryID]tiff.Group{
		model: tiff.GroupIfd0, // Model belongs to the first IFD (aka IFD#0)
		// ...
	})

	// provide the IDs of the entries you would like to collect
	entries, err := p.Parse(tiff.ImageWidth, model)
	if err != nil {
		panic(err)
	}

	if en, ok := entries[tiff.ImageWidth]; ok {
		// read the value, casting it to the expected data type
		width, err := en.ReadUint16()
		if err != nil {
			panic(err)
		}
		fmt.Println("width", width)
	}

	if en, ok := entries[model]; ok {
		model, err := en.ReadString(p)
		if err != nil {
			panic(err)
		}
		fmt.Println("model", model)
	}
}
