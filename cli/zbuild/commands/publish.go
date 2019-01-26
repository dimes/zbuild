package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/dimes/zbuild/artifacts"
	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/local"
	"github.com/dimes/zbuild/model"
)

type publish struct{}

func (p *publish) Describe() string {
	return "Publishes a package in a source set"
}

func (p *publish) Exec(workingDir string, args ...string) error {
	parsedBuildfile, err := model.ParseBuildfile(filepath.Join(workingDir, model.BuildfileName))
	if err != nil {
		buildlog.Fatalf("Error parsing buildfile: %+v", err)
	}

	buildlog.Infof("Registering %s from %s", parsedBuildfile.String(), workingDir)
	localManager, err := local.NewLocalManager(workingDir)
	if err != nil {
		return fmt.Errorf("Error getting local manager for %s: %+v", workingDir, err)
	}

	remoteManager, err := local.GetRemoteManager(workingDir)
	if err != nil {
		return fmt.Errorf("Error getting remote manager for %s: %+v", workingDir, err)
	}

	buildNumber := fmt.Sprintf("%d", time.Now().Unix())
	artifact := model.NewArtifact(parsedBuildfile.Package, buildNumber)

	if err := artifacts.Transfer(localManager, remoteManager, artifact); err != nil {
		return fmt.Errorf("Error transfering %s: %+v", artifact.String(), err)
	}

	remoteSourceSet, err := local.GetRemoteSourceSet(workingDir)
	if err != nil {
		return fmt.Errorf("Error getting remote source set for %s: %+v", workingDir, err)
	}

	if err := remoteSourceSet.RegisterArtifact(artifact); err != nil {
		return fmt.Errorf("Error registering artifact: %+v", err)
	}

	if err := remoteSourceSet.UseArtifact(artifact); err != nil {
		return fmt.Errorf("Error using artifact in source set: %+v", err)
	}

	return nil
}
