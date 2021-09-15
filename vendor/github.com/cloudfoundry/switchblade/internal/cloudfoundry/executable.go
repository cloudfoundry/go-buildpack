package cloudfoundry

import "github.com/paketo-buildpacks/packit/pexec"

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}
