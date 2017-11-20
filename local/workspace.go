package local

import (
	"builder/artifacts"
	"builder/buildlog"
	"builder/model"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// WorkspaceMetadata represents local workspace metadata
type WorkspaceMetadata struct {
	SourceSetName string
	Artifacts     []*model.Artifact
}

const (
	rootSep          = string(os.PathSeparator)
	workspaceDirName = ".workspace"
	metadataFileName = "metadata.json"
)

var (
	// ErrWorkspaceNotFound is returned when no workspace is found
	ErrWorkspaceNotFound = errors.New("workspace not found")
)

// InitWorkspace creates a new workspace at the specified location
func InitWorkspace(location string, sourceSet artifacts.SourceSet) error {
	workspaceDir := filepath.Join(location, workspaceDirName)
	if info, _ := os.Stat(workspaceDir); info != nil {
		return fmt.Errorf("Found existing workspace directory at %s", workspaceDir)
	}

	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return fmt.Errorf("Error creating workspace directory at %s: %+v", workspaceDir, err)
	}

	return RefreshWorkspace(location, sourceSet)
}

// RefreshWorkspace refreshes the workspace metadata for the workspace located at location
func RefreshWorkspace(location string, sourceSet artifacts.SourceSet) error {
	artifacts, err := sourceSet.GetAllArtifacts()
	if err != nil {
		return fmt.Errorf("Error retrieving artifacts from source set %s: %+v", sourceSet.Name(), err)
	}

	workspaceMetadata := &WorkspaceMetadata{
		SourceSetName: sourceSet.Name(),
		Artifacts:     artifacts,
	}

	workspaceDir := filepath.Join(location, workspaceDirName)
	metadataFileLocation := filepath.Join(workspaceDir, metadataFileName)
	openFlags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY

	metadataFile, err := os.OpenFile(metadataFileLocation, openFlags, 0644)
	if err != nil {
		return fmt.Errorf("Error creating metadata file in %s: %+v", metadataFileLocation, err)
	}
	defer metadataFile.Close()

	if err := json.NewEncoder(metadataFile).Encode(workspaceMetadata); err != nil {
		return fmt.Errorf("Error writing workspace metadata to %s: %+v", metadataFileLocation, err)
	}

	return nil
}

// GetWorkspace traverses up the directory tree looking for the workspace directory.
func GetWorkspace(directory string) (string, error) {
	abs, err := filepath.Abs(directory)
	if err != nil {
		return "", fmt.Errorf("Error determining absolute path for %s: %+v", directory, err)
	}

	needToCheckRoot := true
	for dir := abs; dir != rootSep || needToCheckRoot; dir = filepath.Dir(dir) {
		if dir == rootSep {
			needToCheckRoot = false
		}

		metadataFileLocation := filepath.Join(dir, workspaceDirName, metadataFileName)
		info, err := os.Stat(metadataFileLocation)
		if err != nil {
			buildlog.Debugf("Did not find workspace in %s", dir)
			continue
		}

		if info.IsDir() {
			return "", fmt.Errorf("Expected metadata file at %s but found directory", metadataFileLocation)
		}

		return dir, nil
	}

	return "", ErrWorkspaceNotFound
}

// GetWorkspaceMetadata returns the workspace metadata for the workspace containing the
// directory. Returns ErrWorkspaceNotFound if no workspace is found
func GetWorkspaceMetadata(directory string) (*WorkspaceMetadata, error) {
	workspace, err := GetWorkspace(directory)
	if err != nil {
		return nil, err
	}

	metadataFileLocation := filepath.Join(workspace, workspaceDirName, metadataFileName)
	metadataFile, err := os.Open(metadataFileLocation)
	if err != nil {
		return nil, fmt.Errorf("Error opening workspace metadata in %s: %+v", metadataFileLocation, err)
	}
	defer metadataFile.Close()

	workspaceMetadata := &WorkspaceMetadata{}
	if err = json.NewDecoder(metadataFile).Decode(workspaceMetadata); err != nil {
		return nil, fmt.Errorf("Error decoding metadata: %+v", err)
	}

	return workspaceMetadata, err
}
