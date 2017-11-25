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

// GCSManager stores artifacts in GCS
type GCSManager struct {
}
