package tiff

// wanted represents a set of Entry IDs
type wanted struct {
	ids map[EntryID]struct{}
	max EntryID
}

func newWanted(ids ...EntryID) *wanted {
	we := wanted{
		ids: map[EntryID]struct{}{},
	}
	for _, id := range ids {
		we.Put(id)
	}

	return &we
}

func (we *wanted) Put(id EntryID) {
	we.ids[id] = struct{}{}
	if id > we.max {
		we.max = id
	}
}

func (we *wanted) Contains(id EntryID) bool {
	_, ok := we.ids[id]
	return ok
}

func (we *wanted) Max() EntryID {
	return we.max
}

func (we *wanted) Empty() bool {
	return len(we.ids) == 0
}
