package main

import (
	"encoding/json"
	"golang"
	"golang/finalize"
	_ "golang/hooks"
	"os"

	"github.com/cloudfoundry/libbuildpack"
)

func main() {
	stager, err := libbuildpack.NewStager(os.Args[1:], libbuildpack.NewLogger())

	err = libbuildpack.SetStagingEnvironment(stager.DepsDir)
	if err != nil {
		stager.Log.Error("Unable to setup environment variables: %s", err.Error())
		os.Exit(10)
	}

	var godep golang.Godep

	if os.Getenv("supply_VendorTool") == "godep" {
		if err := json.Unmarshal([]byte(os.Getenv("supply_Godep")), &godep); err != nil {
			stager.Log.Error("Unable to load supply_Godep json: %s", err.Error())
			os.Exit(11)
		}
	}

	gf := finalize.Finalizer{
		Stager:     stager,
		Godep:      godep,
		GoVersion:  os.Getenv("supply_GoVersion"),
		VendorTool: os.Getenv("supply_VendorTool"),
	}

	err = finalize.Run(&gf)
	if err != nil {
		os.Exit(12)
	}

	err = libbuildpack.SetLaunchEnvironment(stager.DepsDir, stager.BuildDir)
	if err != nil {
		stager.Log.Error("Unable to setup launch environment: %s", err.Error())
		os.Exit(13)
	}

	err = libbuildpack.RunAfterCompile(stager)
	if err != nil {
		stager.Log.Error("After Compile: %s", err.Error())
		os.Exit(14)
	}

	stager.StagingComplete()
}
