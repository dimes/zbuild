package commands

import (
	"builder"
	"builder/buildlog"
	"builder/gobuilder"
	"builder/model"
	"path/filepath"
)

type build struct{}

func (b *build) Exec(workingDir string, args ...string) error {
	builder.RegisterBuilder(gobuilder.NewGoBuilder())

	parsedBuildfile, err := model.ParseBuildfile(filepath.Join(workingDir, model.BuildfileName))
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

	return nil
}
