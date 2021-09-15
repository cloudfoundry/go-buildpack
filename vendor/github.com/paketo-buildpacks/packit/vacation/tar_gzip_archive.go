package vacation

import (
	"compress/gzip"
	"fmt"
	"io"
)

// A TarGzipArchive decompresses gziped tar files from an input stream.
type TarGzipArchive struct {
	reader     io.Reader
	components int
}

// NewTarGzipArchive returns a new TarGzipArchive that reads from inputReader.
func NewTarGzipArchive(inputReader io.Reader) TarGzipArchive {
	return TarGzipArchive{reader: inputReader}
}

// Decompress reads from TarGzipArchive and writes files into the destination
// specified.
func (gz TarGzipArchive) Decompress(destination string) error {
	gzr, err := gzip.NewReader(gz.reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}

	return NewTarArchive(gzr).StripComponents(gz.components).Decompress(destination)
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (gz TarGzipArchive) StripComponents(components int) TarGzipArchive {
	gz.components = components
	return gz
}
