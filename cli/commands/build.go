package commands

import (
	"path/filepath"

	"github.com/dimes/zbuild"
	"github.com/dimes/zbuild/builders/golang"
	"github.com/dimes/zbuild/builders/protobuf"
	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/model"
)

type build struct{}

func (b *build) Describe() string {
	return "Builds a package"
}

func (b *build) Exec(workingDir string, args ...string) error {
	zbuild.RegisterBuilder(golang.NewBuilder())
	zbuild.RegisterBuilder(protobuf.NewBuilder())

	parsedBuildfile, err := model.ParseBuildfile(filepath.Join(workingDir, model.BuildfileName))
	if err != nil {
		buildlog.Fatalf("Error parsing buildfile: %+v", err)
	}

	buildlog.Infof("Parsed buildfile for %s", parsedBuildfile.Package.String())
	builder := zbuild.GetBuilderForType(parsedBuildfile.Type)
	if builder == nil {
		buildlog.Fatalf("Could not find builder for type %s", parsedBuildfile.Type)
	}

	if err = builder.Build(parsedBuildfile); err != nil {
		buildlog.Fatalf("Error during build: %+v", err)
	}

	return nil
}
