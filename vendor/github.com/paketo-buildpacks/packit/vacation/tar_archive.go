package vacation

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// A TarArchive decompresses tar files from an input stream.
type TarArchive struct {
	reader     io.Reader
	components int
}

// NewTarArchive returns a new TarArchive that reads from inputReader.
func NewTarArchive(inputReader io.Reader) TarArchive {
	return TarArchive{reader: inputReader}
}

// Decompress reads from TarArchive and writes files into the
// destination specified.
func (ta TarArchive) Decompress(destination string) error {
	// This map keeps track of what directories have been made already so that we
	// only attempt to make them once for a cleaner interaction.  This map is
	// only necessary in cases where there are no directory headers in the
	// tarball, which can be seen in the test around there being no directory
	// metadata.
	directories := map[string]interface{}{}

	// Struct and slice to collect symlinks and create them after all files have
	// been created
	type header struct {
		name     string
		linkname string
		path     string
	}

	var symlinkHeaders []header

	tarReader := tar.NewReader(ta.reader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar response: %s", err)
		}

		// Clean the name in the header to prevent './filename' being stripped to
		// 'filename' also to skip if the destination it the destination directory
		// itself i.e. './'
		var name string
		if name = filepath.Clean(hdr.Name); name == "." {
			continue
		}

		err = checkExtractPath(name, destination)
		if err != nil {
			return err
		}

		fileNames := strings.Split(name, "/")

		// Checks to see if file should be written when stripping components
		if len(fileNames) <= ta.components {
			continue
		}

		// Constructs the path that conforms to the stripped components.
		path := filepath.Join(append([]string{destination}, fileNames[ta.components:]...)...)

		// This switch case handles all cases for creating the directory structure
		// this logic is needed to handle tarballs with no directory headers.
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create archived directory: %s", err)
			}

			directories[path] = nil

		default:
			dir := filepath.Dir(path)
			_, ok := directories[dir]
			if !ok {
				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					return fmt.Errorf("failed to create archived directory from file path: %s", err)
				}
				directories[dir] = nil
			}
		}

		// This switch case handles the creation of files during the untaring process.
		switch hdr.Typeflag {
		case tar.TypeReg:
			file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create archived file: %s", err)
			}

			_, err = io.Copy(file, tarReader)
			if err != nil {
				return err
			}

			err = file.Close()
			if err != nil {
				return err
			}

		case tar.TypeSymlink:
			// Collect all of the headers for symlinks so that they can be verified
			// after all other files are written
			symlinkHeaders = append(symlinkHeaders, header{
				name:     hdr.Name,
				linkname: hdr.Linkname,
				path:     path,
			})
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
			return fmt.Errorf("failed to extract symlink: %s", err)
		}
	}

	return nil
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (ta TarArchive) StripComponents(components int) TarArchive {
	ta.components = components
	return ta
}
