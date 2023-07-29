package tiff

import "fmt"

type ErrNotFound struct {
	message string
}

func (e ErrNotFound) Error() string {
	return e.message
}

func notFound(ext string) ErrNotFound {
	return ErrNotFound{message: fmt.Sprintf("not found: unknown extension '%s' or no exif data", ext)}
}
