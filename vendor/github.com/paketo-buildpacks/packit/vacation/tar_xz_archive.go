package vacation

import (
	"fmt"
	"io"

	"github.com/ulikunitz/xz"
)

// A TarXZArchive decompresses xz tar files from an input stream.
type TarXZArchive struct {
	reader     io.Reader
	components int
}

// NewTarXZArchive returns a new TarXZArchive that reads from inputReader.
func NewTarXZArchive(inputReader io.Reader) TarXZArchive {
	return TarXZArchive{reader: inputReader}
}

// Decompress reads from TarXZArchive and writes files into the destination
// specified.
func (txz TarXZArchive) Decompress(destination string) error {
	xzr, err := xz.NewReader(txz.reader)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	return NewTarArchive(xzr).StripComponents(txz.components).Decompress(destination)
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (txz TarXZArchive) StripComponents(components int) TarXZArchive {
	txz.components = components
	return txz
}
