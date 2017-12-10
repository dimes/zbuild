// Package artifacts contains definitions required for artifact manipulation
package artifacts

import (
	"fmt"
	"io"

	"github.com/dimes/zbuild/model"
)

// Transfer transfers an artifact from source to the destination. Note: This does not explicitly
// update the source set.
func Transfer(source Manager, destination Manager, artifact *model.Artifact) error {
	reader, err := source.OpenReader(artifact)
	if err != nil {
		return fmt.Errorf("Error opening reader to source for %s: %+v", artifact.String(), err)
	}
	defer reader.Close()

	writer, err := source.OpenWriter(artifact)
	if err != nil {
		return fmt.Errorf("Error opening writer to destination for %s: %+v", artifact.String(), err)
	}
	defer writer.Close()

	if _, err = io.Copy(writer, reader); err != nil {
		return fmt.Errorf("Error copying source to destination for %s: %+v", artifact.String(), err)
	}

	return nil
}
