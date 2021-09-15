package vacation

import (
	"io"
	"os"
)

// A NopArchive implements the common archive interface, but acts as a no-op,
// simply copying the reader to the destination.
type NopArchive struct {
	reader io.Reader
}

// NewNopArchive returns a new NopArchive
func NewNopArchive(r io.Reader) NopArchive {
	return NopArchive{reader: r}
}

// Decompress copies the reader contents into the destination specified.
func (na NopArchive) Decompress(destination string) error {
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, na.reader)
	if err != nil {
		return err
	}

	return nil
}
