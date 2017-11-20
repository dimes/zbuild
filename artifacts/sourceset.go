package artifacts

import "builder/model"

// SourceSet represents a set of packages. This set of packages are used to resolve package dependencies.
// Implementations will typically rely on a notion of a "workspace" that contains packages as well as
// metadata about the set. The set of packages in a source set are represented by "Artifacts". The
// artfacts represent specific builds of the constituent packages.
type SourceSet interface {
	Name() string
	GetArtifact(namespace, name, version string) (*model.Artifact, error)
	GetAllArtifacts() ([]*model.Artifact, error)
}

// DynamoSourceSet uses DynamoDB to store package information
type DynamoSourceSet struct {
}

// DatastoreSourceSet uses Datastore to store package information
type DatastoreSourceSet struct {
}
