package main

import (
	"golang/finalize"
	_ "golang/hooks"
	"os"

	"github.com/cloudfoundry/libbuildpack"
)

type config struct {
  Config struct {
    GoVersion  string `yaml:"GoVersion"`
    VendorTool string `yaml:"VendorTool"`
    Godep      string `yaml:"Godep"`
  } `yaml:"config"`
}

func main() {
	stager, err := libbuildpack.NewStager(os.Args[1:], libbuildpack.NewLogger())

	if err := libbuildpack.SetStagingEnvironment(stager.DepsDir); err != nil {
		stager.Log.Error("Unable to setup environment variables: %s", err.Error())
		os.Exit(10)
	}

	gf, err := finalize.NewFinalizer(stager)
	if err != nil {
		os.Exit(11)
	}

	if err := finalize.Run(gf); err != nil {
		os.Exit(12)
	}

	if err := libbuildpack.SetLaunchEnvironment(stager.DepsDir, stager.BuildDir); err != nil {
		stager.Log.Error("Unable to setup launch environment: %s", err.Error())
		os.Exit(13)
	}

	if err := libbuildpack.RunAfterCompile(stager); err != nil {
		stager.Log.Error("After Compile: %s", err.Error())
		os.Exit(14)
	}

	stager.StagingComplete()
}
