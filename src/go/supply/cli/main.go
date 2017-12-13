package main

import (
	_ "go/hooks"
	"go/supply"
	"os"
	"time"

	"github.com/SUSE/cf-libbuildpack"
)

func main() {
	logger := libbuildpack.NewLogger(os.Stdout)

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		logger.Error("Unable to determine buildpack directory: %s", err.Error())
		os.Exit(8)
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		logger.Error("Unable to load buildpack manifest: %s", err.Error())
		os.Exit(9)
	}

	stager := libbuildpack.NewStager(os.Args[1:], logger, manifest)
	if err = stager.CheckBuildpackValid(); err != nil {
		os.Exit(10)
	}

	if err := libbuildpack.RunBeforeCompile(stager); err != nil {
		logger.Error("Before Compile: %s", err.Error())
		os.Exit(12)
	}

	if err := stager.SetStagingEnvironment(); err != nil {
		logger.Error("Unable to setup environment variables: %s", err.Error())
		os.Exit(13)
	}

	gs := supply.Supplier{
		Stager:   stager,
		Log:      logger,
		Manifest: manifest,
	}

	if err := supply.Run(&gs); err != nil {
		os.Exit(16)
	}
}
