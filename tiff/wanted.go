package tiff

// wanted represents a set of Entry IDs
type wanted struct {
	ids map[uint16]struct{}
	max uint16
}

func newWanted(ids ...uint16) *wanted {
	we := wanted{
		ids: map[uint16]struct{}{},
	}
	for _, id := range ids {
		we.Put(id)
	}

	return &we
}

func (we *wanted) Put(id uint16) {
	we.ids[id] = struct{}{}
	if id > we.max {
		we.max = id
	}
}

func (we *wanted) Contains(id uint16) bool {
	_, ok := we.ids[id]
	return ok
}

func (we *wanted) Max() uint16 {
	return we.max
}
