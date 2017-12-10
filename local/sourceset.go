package local

import (
	"fmt"
	"github.com/dimes/zbuild/artifacts"
	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/model"
	"io"
	"io/ioutil"
	"path/filepath"
)

type localSourceSet struct {
	workspace     string
	name          string
	artifacts     []*model.Artifact
	artifactIndex map[string]map[string]map[string]*model.Artifact
}

type overrideSourceSet struct {
	localSourceSet
	overrideLocations map[string]string
}

// NewLocalSourceSet returns a source set that uses the workspace directory as its backing data.
// Any directory inside the workspace can be passed as this directory string
func NewLocalSourceSet(directory string) (artifacts.SourceSet, error) {
	workspace, err := GetWorkspace(directory)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace directory for %s: %+v", directory, err)
	}

	workspaceMetadata, err := GetWorkspaceMetadata(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace metadata for source set: %+v", err)
	}

	return newLocalSourceSet(workspace, workspaceMetadata.SourceSetName, workspaceMetadata.Artifacts), nil
}

// NewOverrideSourceSet returns a source set that uses only packages checked out in the workspace
func newOverrideSourceSet(directory string) (*overrideSourceSet, error) {
	workspace, err := GetWorkspace(directory)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace directory for %s: %+v", directory, err)
	}

	files, err := ioutil.ReadDir(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error listing workspace %s: %+v", workspace, err)
	}

	artifacts := make([]*model.Artifact, 0)
	overrideLocations := make(map[string]string)
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		packagePath := filepath.Join(workspace, file.Name())
		buildfilePath := filepath.Join(packagePath, model.BuildfileName)
		parsedBuildfile, err := model.ParseBuildfile(buildfilePath)
		if err != nil {
			buildlog.Debugf("Ignoring possible override %s: %+v", buildfilePath, err)
			continue
		}

		artifacts = append(artifacts, &model.Artifact{
			Package: parsedBuildfile.Package,
		})
		overrideLocations[packageToMapKey(parsedBuildfile.Package)] = packagePath
	}

	workspaceMetadata, err := GetWorkspaceMetadata(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace metadata for %s: %+v", directory, err)
	}

	localSourceSet := newLocalSourceSet(workspace, workspaceMetadata.SourceSetName, artifacts)
	return &overrideSourceSet{
		localSourceSet:    *localSourceSet,
		overrideLocations: overrideLocations,
	}, nil
}

func newLocalSourceSet(workspace, name string, artifacts []*model.Artifact) *localSourceSet {
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
		workspace:     workspace,
		name:          name,
		artifacts:     artifacts,
		artifactIndex: artifactIndex,
	}
}

func (l *localSourceSet) Type() string {
	return "local"
}

func (l *localSourceSet) Setup() error {
	return nil
}

func (l *localSourceSet) Name() string {
	return l.name
}

func (l *localSourceSet) GetArtifact(namespace, name, version string) (*model.Artifact, error) {
	namespaceArtifacts, ok := l.artifactIndex[namespace]
	if !ok {
		buildlog.Debugf("Namespace %s not found", namespace)
		return nil, artifacts.ErrArtifactNotFound

	}

	versions, ok := namespaceArtifacts[name]
	if !ok {
		buildlog.Debugf("Package %s/%s not found in workspace", namespace, name)
		return nil, artifacts.ErrArtifactNotFound
	}

	artifact, ok := versions[version]
	if !ok {
		buildlog.Debugf("Version %s not found for %s/%s in workspace", version, namespace, name)
		return nil, artifacts.ErrArtifactNotFound
	}

	return artifact, nil
}

func (l *localSourceSet) GetAllArtifacts() ([]*model.Artifact, error) {
	return l.artifacts, nil
}

func (o *overrideSourceSet) getLocationForArtifact(namespace, name, version string) (string, error) {
	if override := o.overrideLocations[packageInfoToMapKey(namespace, name, version)]; override != "" {
		return override, nil
	}

	return "", fmt.Errorf("No override location found for %s/%s/%s", namespace, name, version)
}

func (l *localSourceSet) PersistMetadata(writer io.Writer) error {
	return nil
}
