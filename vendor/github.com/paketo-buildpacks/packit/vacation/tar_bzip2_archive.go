package vacation

import (
	"compress/bzip2"
	"io"
)

// A TarBzip2Archive decompresses bzip2 files from an input stream.
type TarBzip2Archive struct {
	reader     io.Reader
	components int
}

// NewTarBzip2Archive returns a new Bzip2Archive that reads from inputReader.
func NewTarBzip2Archive(inputReader io.Reader) TarBzip2Archive {
	return TarBzip2Archive{reader: inputReader}
}

// Decompress reads from TarBzip2Archive and writes files into the destination
// specified.
func (tbz TarBzip2Archive) Decompress(destination string) error {
	return NewTarArchive(bzip2.NewReader(tbz.reader)).StripComponents(tbz.components).Decompress(destination)
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (tbz TarBzip2Archive) StripComponents(components int) TarBzip2Archive {
	tbz.components = components
	return tbz
}
