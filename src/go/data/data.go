package data

import (
	"fmt"
	"path"
)

func ReleaseYAML(mainPackageName string) string {
	release := `---
default_process_types:
    web: ./bin/%s
`
	return fmt.Sprintf(release, path.Base(mainPackageName))
}

func GoScript() string {
	return "PATH=$PATH:$HOME/bin\n"
}

func GoRootScript(goRoot string) string {
	contents := `export GOROOT=%s
PATH=$PATH:$GOROOT/bin
`

	return fmt.Sprintf(contents, goRoot)
}

func ZZGoPathScript(mainPackageName string) string {
	contents := `export GOPATH=$HOME
cd $GOPATH/src/%s
`
	return fmt.Sprintf(contents, path.Base(mainPackageName))
}
