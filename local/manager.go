package local

import (
	"archive/tar"
	"builder/artifacts"
	"builder/buildlog"
	"builder/model"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// LocalManagerType is the type identifier for local managers.
	LocalManagerType = "local"

	workspacePackageCacheDirName = "package-cache"
)

type localManager struct {
	workspace string
}

// NewLocalManager returns a manager that can be used for reading / writing local artifacts.
// The directory can be any directory in a workspace
//
// The local manager is designed to work with a type of remote manager, which manages artifacts in the form
// of tarballs. Therefore, opening a reader to a local artifact will read a tar file and opening a
// writer to a local workspace will assume the input is a tarball.
//
// Furthermore, a reader will assume to be reading a local package the user has checked out, but a writer
// is assumed to be writing to the local package cache in the workspace metadata. These considerations
// arise from practical considerations: a user wants to download packages locally for building, and then
// want to possibly upload local packages once they've made their changes.
func NewLocalManager(directory string) (artifacts.Manager, error) {
	workspace, err := GetWorkspace(directory)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace for %s: %+v", directory, err)
	}

	return &localManager{
		workspace: workspace,
	}, nil
}

func (l *localManager) Type() string {
	return LocalManagerType
}

func (l *localManager) Setup() error {
	return nil
}

// OpenReader opens a reader to the given artifact. Because the artifact is local, the "build number"
// is ignored
func (l *localManager) OpenReader(artifact *model.Artifact) (io.ReadCloser, error) {
	// The artifact needs to be found in the metadata directory
	dirs, err := ioutil.ReadDir(l.workspace)
	if err != nil {
		return nil, fmt.Errorf("Error listing files in workspace %s: %+v", l.workspace, err)
	}

	packageDir := ""
	for _, dir := range dirs {
		if dir.Name() == workspaceDirName {
			buildlog.Debugf("Ignoring %s because it contains workspace metadata", workspaceDirName)
			continue
		}

		if !dir.IsDir() {
			buildlog.Debugf("Ignoring %s because it is not a directory", dir.Name())
			continue
		}

		buildfileLocation := filepath.Join(l.workspace, dir.Name(), model.BuildfileName)
		parsedBuildfile, err := model.ParseBuildfile(buildfileLocation)
		if err != nil {
			buildlog.Debugf("Ignoring %s due to error opening build file: %+v", dir.Name(), err)
			continue
		}

		if parsedBuildfile.Namespace != artifact.Namespace ||
			parsedBuildfile.Name != artifact.Name ||
			parsedBuildfile.Version != artifact.Version {
			buildlog.Debugf("Ignoring %s due to namespace/name/version mismatch", dir.Name())
			continue
		}

		if packageDir != "" {
			return nil, fmt.Errorf("Found duplicate workspace packages for %+v. %s and %s", artifact,
				packageDir, filepath.Join(l.workspace, dir.Name()))
		}

		packageDir = filepath.Join(l.workspace, dir.Name())
	}

	if packageDir == "" {
		return nil, fmt.Errorf("Could not find local package matching artifact %+v", artifact)
	}

	// We only want to package up the build directory
	packageDir = filepath.Join(packageDir, model.BuildDir)

	reader, writer := io.Pipe()
	go func() {
		gzipWriter := gzip.NewWriter(writer)
		tarWriter := tar.NewWriter(gzipWriter)
		defer writer.Close()
		defer gzipWriter.Close()
		defer tarWriter.Close()

		err := filepath.Walk(packageDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if path == packageDir {
				return nil
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return fmt.Errorf("Error creating file info header for %s: %+v", path, err)
			}

			name, err := filepath.Rel(packageDir, path)
			if err != nil {
				return fmt.Errorf("Error getting %s relative to %s: %+v", path, packageDir, err)
			}
			header.Name = name

			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("Error writing header for %s: %+v", path, err)
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("Error opening %s: %+v", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("Error copying %s: %+v", path, err)
			}

			return nil
		})

		if err != nil {
			writer.CloseWithError(err)
		}
	}()

	return reader, nil
}

// Opens a writer for the given artifact into the workspace package cache
func (l *localManager) OpenWriter(artifact *model.Artifact) (io.WriteCloser, error) {
	artifactDirName := localArtifactCacheDir(l.workspace, artifact)
	if err := os.RemoveAll(artifactDirName); err != nil {
		return nil, fmt.Errorf("Error removing existing artifact at %s: %+v", artifactDirName, err)
	}

	if err := os.MkdirAll(artifactDirName, 0755); err != nil {
		return nil, fmt.Errorf("Error creating artifact director %s: %+v", artifactDirName, err)
	}

	reader, writer := io.Pipe()
	go func() {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			reader.CloseWithError(fmt.Errorf("Error opening gzip reader: %+v", err))
		}

		defer reader.Close()
		defer gzipReader.Close()

		tarReader := tar.NewReader(gzipReader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				return
			}

			if err != nil {
				reader.CloseWithError(fmt.Errorf("Error reading tar header for %s: %+v", artifactDirName, err))
				return
			}

			if header == nil {
				continue
			}

			destination := filepath.Join(artifactDirName, header.Name)
			if header.Typeflag == tar.TypeDir {
				if err := os.MkdirAll(destination, 0755); err != nil {
					reader.CloseWithError(fmt.Errorf("Error creating directory %s: %+v", destination, err))
					return
				}
				continue
			} else if header.Typeflag == tar.TypeReg {
				flags := os.O_CREATE | os.O_EXCL | os.O_WRONLY
				file, err := os.OpenFile(destination, flags, os.FileMode(header.Mode))
				if err != nil {
					reader.CloseWithError(fmt.Errorf("Error opening file %s: %+v", destination, err))
					return
				}

				if _, err := io.Copy(file, tarReader); err != nil {
					reader.CloseWithError(fmt.Errorf("Error copying file %s: %+v", destination, err))
					file.Close()
					return
				}
				file.Close()
			} else {
				buildlog.Debugf("Unknown header typeflag %.2x", header.Typeflag)
			}
		}
	}()

	return writer, nil
}

func (l *localManager) PersistMetadata(writer io.Writer) error {
	return nil
}

func localArtifactCacheDir(workspace string, artifact *model.Artifact) string {
	return filepath.Join(workspace, workspaceDirName, workspacePackageCacheDirName,
		artifact.Namespace, artifact.Name, artifact.Version, artifact.BuildNumber)
}

// GetRemoteManager returns the manager configured for the workspace directory
func GetRemoteManager(directory string) (artifacts.Manager, error) {
	workspace, err := GetWorkspace(directory)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace for %s: %+v", directory, err)
	}

	workspaceMetadata, err := GetWorkspaceMetadata(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace metadata for %s: %+v", workspace, err)
	}

	managerMetadataFile, err := os.Open(filepath.Join(workspace, workspaceDirName, managerFileName))
	if err != nil {
		return nil, fmt.Errorf("Error opening manager metadata file: %+v", err)
	}
	defer managerMetadataFile.Close()

	switch workspaceMetadata.ManagerType {
	case artifacts.S3ManagerType:
		metadata := &artifacts.S3Metadata{}
		if err := json.NewDecoder(managerMetadataFile).Decode(metadata); err != nil {
			return nil, fmt.Errorf("Error decoding manager metadata: %+v", err)
		}
		return artifacts.NewS3ManagerFromMetadata(s3.New(NewSession("", metadata.Profile)), metadata)
	default:
		return nil, fmt.Errorf("Unknown manager type found in metadata: %s", workspaceMetadata.ManagerType)
	}
}
