package artifacts

import (
	"builder/model"
	"io"
)

// Manager implementations can read/write artifacts to backing data store
type Manager interface {
	Type() string
	Setup() error // Idempotently creates any necessary structures for the manager, e.g. Dynamo tables
	OpenReader(artifact *model.Artifact) (io.ReadCloser, error)
	OpenWriter(artifact *model.Artifact) (io.WriteCloser, error)
	PersistMetadata(writer io.Writer) error
}

// GCSManager stores artifacts in GCS
type GCSManager struct {
}
