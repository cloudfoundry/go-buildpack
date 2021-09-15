package vacation

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// A ZipArchive decompresses zip files from an input stream.
type ZipArchive struct {
	reader     io.Reader
	components int
}

// NewZipArchive returns a new ZipArchive that reads from inputReader.
func NewZipArchive(inputReader io.Reader) ZipArchive {
	return ZipArchive{reader: inputReader}
}

// Decompress reads from ZipArchive and writes files into the destination
// specified.
func (z ZipArchive) Decompress(destination string) error {
	// Struct and slice to collect symlinks and create them after all files have
	// been created
	type header struct {
		name     string
		linkname string
		path     string
	}

	var symlinkHeaders []header

	// Use an os.File to buffer the zip contents. This is needed because
	// zip.NewReader requires an io.ReaderAt so that it can jump around within
	// the file as it decompresses.
	buffer, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer os.Remove(buffer.Name())

	size, err := io.Copy(buffer, z.reader)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(buffer, size)
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, f := range zr.File {
		// Clean the name in the header to prevent './filename' being stripped to
		// 'filename' also to skip if the destination it the destination directory
		// itself i.e. './'
		var name string
		if name = filepath.Clean(f.Name); name == "." {
			continue
		}

		err = checkExtractPath(name, destination)
		if err != nil {
			return err
		}

		fileNames := strings.Split(name, "/")

		// Checks to see if file should be written when stripping components
		if len(fileNames) <= z.components {
			continue
		}

		// Constructs the path that conforms to the stripped components.
		path := filepath.Join(append([]string{destination}, fileNames[z.components:]...)...)

		switch {
		case f.FileInfo().IsDir():
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to unzip directory: %w", err)
			}
		case f.FileInfo().Mode()&os.ModeSymlink != 0:
			fd, err := f.Open()
			if err != nil {
				return err
			}

			linkname, err := io.ReadAll(fd)
			if err != nil {
				return err
			}

			// Collect all of the headers for symlinks so that they can be verified
			// after all other files are written
			symlinkHeaders = append(symlinkHeaders, header{
				name:     f.Name,
				linkname: string(linkname),
				path:     path,
			})

		default:
			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to unzip directory that was part of file path: %w", err)
			}

			dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("failed to unzip file: %w", err)
			}

			src, err := f.Open()
			if err != nil {
				return err
			}

			_, err = io.Copy(dst, src)
			if err != nil {
				return err
			}

			if err := dst.Close(); err != nil {
				return err
			}

			if err := src.Close(); err != nil {
				return err
			}
		}
	}

	// Sort the symlinks so that symlinks of symlinks have their base link
	// created before they are created.
	//
	// For example:
	// b-sym -> a-sym/x
	// a-sym -> z
	// c-sym -> d-sym
	// d-sym -> z
	//
	// Will sort to:
	// a-sym -> z
	// b-sym -> a-sym/x
	// d-sym -> z
	// c-sym -> d-sym
	sort.Slice(symlinkHeaders, func(i, j int) bool {
		if filepath.Clean(symlinkHeaders[i].name) == linknameFullPath(symlinkHeaders[j].name, symlinkHeaders[j].linkname) {
			return true
		}

		if filepath.Clean(symlinkHeaders[j].name) == linknameFullPath(symlinkHeaders[i].name, symlinkHeaders[i].linkname) {
			return false
		}

		return filepath.Clean(symlinkHeaders[i].name) < linknameFullPath(symlinkHeaders[j].name, symlinkHeaders[j].linkname)
	})

	for _, h := range symlinkHeaders {
		// Check to see if the file that will be linked to is valid for symlinking
		_, err := filepath.EvalSymlinks(linknameFullPath(h.path, h.linkname))
		if err != nil {
			return fmt.Errorf("failed to evaluate symlink %s: %w", h.path, err)
		}

		err = os.Symlink(h.linkname, h.path)
		if err != nil {
			return fmt.Errorf("failed to unzip symlink: %w", err)
		}
	}

	return nil
}

// StripComponents removes the first n levels from the final decompression
// destination.
func (z ZipArchive) StripComponents(components int) ZipArchive {
	z.components = components
	return z
}
