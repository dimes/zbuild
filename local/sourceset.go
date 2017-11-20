package local

import (
	"builder/artifacts"
	"builder/model"
	"fmt"
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

	artifactIndex := make(map[string]map[string]map[string]*model.Artifact)
	for _, artifact := range workspaceMetadata.Artifacts {
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
		name:          workspaceMetadata.SourceSetName,
		artifacts:     workspaceMetadata.Artifacts,
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
