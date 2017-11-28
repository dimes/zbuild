package local

import (
	"builder/artifacts"
	"builder/buildlog"
	"builder/model"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type localSourceSet struct {
	name          string
	artifacts     []*model.Artifact
	artifactIndex map[string]map[string]map[string]*model.Artifact
}

// NewLocalSourceSet returns a source set that uses the workspace directory as its backing data.
// Any directory inside the workspace can be passed as this directory string
func NewLocalSourceSet(directory string) (artifacts.SourceSet, error) {
	workspaceMetadata, err := GetWorkspaceMetadata(directory)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace metadata for source set: %+v", err)
	}

	return newLocalSourceSet(workspaceMetadata.SourceSetName, workspaceMetadata.Artifacts)
}

// NewOverrideSourceSet returns a source set that uses only packages checked out in the workspace
func NewOverrideSourceSet(directory string) (artifacts.SourceSet, error) {
	workspace, err := GetWorkspace(directory)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace directory for %s: %+v", directory, err)
	}

	files, err := ioutil.ReadDir(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error listing workspace %s: %+v", workspace, err)
	}

	artifacts := make([]*model.Artifact, 0)
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		buildfilePath := filepath.Join(workspace, file.Name(), model.BuildfileName)
		parsedBuildfile, err := model.ParseBuildfile(buildfilePath)
		if err != nil {
			buildlog.Debugf("Ignoring possible override %s: %+v", buildfilePath, err)
		}

		artifacts = append(artifacts, &model.Artifact{
			Package: parsedBuildfile.Package,
		})
	}

	workspaceMetadata, err := GetWorkspaceMetadata(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace metadata for %s: %+v", directory, err)
	}

	return newLocalSourceSet(workspaceMetadata.SourceSetName, artifacts)
}

func newLocalSourceSet(name string, artifacts []*model.Artifact) (artifacts.SourceSet, error) {
	artifactIndex := make(map[string]map[string]map[string]*model.Artifact)
	for _, artifact := range artifacts {
		namespace, ok := artifactIndex[artifact.Namespace]
		if !ok {
			namespace = make(map[string]map[string]*model.Artifact)
			artifactIndex[artifact.Namespace] = namespace
		}

		version, ok := namespace[artifact.Name]
		if !ok {
			version = make(map[string]*model.Artifact)
			namespace[artifact.Name] = version
		}

		version[artifact.Version] = artifact
	}

	return &localSourceSet{
		name:          name,
		artifacts:     artifacts,
		artifactIndex: artifactIndex,
	}, nil
}

func (l *localSourceSet) Name() string {
	return l.name
}

func (l *localSourceSet) GetArtifact(namespace, name, version string) (*model.Artifact, error) {
	artifacts, ok := l.artifactIndex[namespace]
	if !ok {
		return nil, fmt.Errorf("Namespace %s not found", namespace)
	}

	versions, ok := artifacts[name]
	if !ok {
		return nil, fmt.Errorf("Package %s/%s not found in workspace", namespace, name)
	}

	artifact, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("Version %s not found for %s/%s in workspace", version, namespace, name)
	}

	return artifact, nil
}

func (l *localSourceSet) GetAllArtifacts() ([]*model.Artifact, error) {
	return l.artifacts, nil
}
