// Package artifacts contains logic for creating and manipulating artifacts.
package artifacts

import (
	"builder/model"
	"io"
)

// Manager implementations can read/write artifacts to backing data store
type Manager interface {
	OpenReader(artifact *model.Artifact) (io.ReadCloser, error)
	OpenWriter(artifact *model.Artifact) (io.WriteCloser, error)
}

// S3Manager stores artifacts in S3
type S3Manager struct {
}

// GCSManager stores artifacts in GCS
type GCSManager struct {
}
