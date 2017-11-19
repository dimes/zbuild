// Package builder contains interfaces and definitions for builders
package builder

import (
	"builder/buildlog"
	"builder/model"
	"fmt"
)

var (
	builders = make(map[string]Builder)
)

// Builder is an interface that all builders must implement
type Builder interface {
	Type() string
	Build(*model.ParsedBuildfile) error
}

// RegisterBuilder associates the given builder with its type. If the type already has
// a builder associated with it, then this method will return an error. This method is not
// safe for concurrent calls
func RegisterBuilder(builder Builder) error {
	if _, ok := builders[builder.Type()]; ok {
		return fmt.Errorf("Type %s is already registered", builder.Type())
	}

	buildlog.Debugf("Associating type %s with %+v", builder.Type(), builder)
	builders[builder.Type()] = builder
	return nil
}

// GetBuilderForType returns a builder for the given type, or nil if no such builder is registered
func GetBuilderForType(builderType string) Builder {
	return builders[builderType]
}
