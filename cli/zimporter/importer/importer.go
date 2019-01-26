// Package importer contains import logic for different languages
package importer

import (
	"github.com/dimes/zbuild/artifacts"
)

// Strategy is an interface that import packages from various languages into
type Strategy interface {
	// Import imports an artifact into the given source set
	Import(sourceSet artifacts.SourceSet, manager artifacts.Manager) error
}

// GoStrategy is a strategy for importing go
type GoStrategy struct {
}
