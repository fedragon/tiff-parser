package tiff

import "github.com/fedragon/tiff-parser/tiff/entry"

// wanted represents a set of Entry IDs
type wanted struct {
	ids map[entry.ID]struct{}
	max entry.ID
}

func newWanted(ids ...entry.ID) *wanted {
	we := wanted{
		ids: map[entry.ID]struct{}{},
	}
	for _, id := range ids {
		we.Put(id)
	}

	return &we
}

func (we *wanted) Put(id entry.ID) {
	we.ids[id] = struct{}{}
	if id > we.max {
		we.max = id
	}
}

func (we *wanted) Contains(id entry.ID) bool {
	_, ok := we.ids[id]
	return ok
}

func (we *wanted) Max() entry.ID {
	return we.max
}

func (we *wanted) Empty() bool {
	return len(we.ids) == 0
}
