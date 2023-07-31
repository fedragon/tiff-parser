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

	// an `entry.ID` is simply a type alias for `uint16`
	model := entry.ID(0x0110)

	// only needed when your entry is not listed in `tiff.Defaults`
	p.WithMapping(map[entry.ID]tiff.Group{
		model: tiff.GroupIfd0, // Model belongs to the first IFD (aka IFD#0)
		// ...
	})

	// provide the IDs of the entries you would like to collect
	entries, err := p.Parse(entry.ImageWidth, model)
	if err != nil {
		panic(err)
	}

	if en, ok := entries[entry.ImageWidth]; ok {
		// read the value, casting it to the expected data type
		width, err := p.ReadUint16(en)
		if err != nil {
			panic(err)
		}
		fmt.Printf("width: %v\n", width)
	}

	if en, ok := entries[model]; ok {
		model, err := p.ReadString(en)
		if err != nil {
			panic(err)
		}
		fmt.Printf("model: %v\n", model)
	}
}
