package gomod

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type GoMod struct {
	GoVersion string
}

// TODO: Replace with https://pkg.go.dev/golang.org/x/mod/modfile when it is available
func (g *GoMod) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open go.mod: %s", err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		text := scanner.Text()
		parts := strings.Split(text, " ")

		if parts[0] == "go" {
			g.GoVersion = parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan go.mod: %s", err)
	}

	return nil
}
