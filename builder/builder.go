package main

import (
	"builder"
	"builder/buildlog"
	"builder/gobuilder"
	"builder/model"
	"flag"
)

const (
	buildfileName = "build.yaml"
)

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "if set, verbose logging will be enabled")
	flag.Parse()

	buildlog.SetLogLevel(buildlog.Info)
	if verbose {
		buildlog.SetLogLevel(buildlog.Debug)
	}

	builder.RegisterBuilder(gobuilder.NewGoBuilder())

	parsedBuildfile, err := model.ParseBuildfile(buildfileName)
	if err != nil {
		buildlog.Fatalf("Error parsing buildfile: %+v", err)
	}

	buildlog.Infof("Parsed buildfile for %s", parsedBuildfile.Package.String())
	builder := builder.GetBuilderForType(parsedBuildfile.Type)
	if builder == nil {
		buildlog.Fatalf("Could not find builder for type %s", parsedBuildfile.Type)
	}

	if err = builder.Build(parsedBuildfile); err != nil {
		buildlog.Fatalf("Error during build: %+v", err)
	}
}
