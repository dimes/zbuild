package artifacts

import (
	"builder/model"
	"errors"
)

var (
	// ErrArtifactNotFound is returned when an artifact is not found
	ErrArtifactNotFound = errors.New("artifact not found")
)

// SourceSet represents a set of packages. This set of packages are used to resolve package dependencies.
// Implementations will typically rely on a notion of a "workspace" that contains packages as well as
// metadata about the set. The set of packages in a source set are represented by "Artifacts". The
// artfacts represent specific builds of the constituent packages.
type SourceSet interface {
	Setup() error
	Name() string
	GetArtifact(namespace, name, version string) (*model.Artifact, error)
	GetAllArtifacts() ([]*model.Artifact, error)
}

// Source Set Table
// source set -> namespace/package-name/version (+ build)

// Package Table (essentially an append only log)
// Package name -> version/build (+ build date)

// DatastoreSourceSet uses Datastore to store package information
type DatastoreSourceSet struct {
}
