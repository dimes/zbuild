// Package model contains the types that represent Builder's data model
package model

import (
	"builder/buildlog"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

const (
	buildDir = "build"
)

// Buildfile is what a package's build file is parsed into
type Buildfile struct {
	Package `yaml:",inline"` // The build file always contains a nested package
}

// ParsedBuildfile is like a Buildfile but contains meta-information about the input build file
type ParsedBuildfile struct {
	Buildfile

	AbsoluteWorkingDir string
	AbsoluteBuildDir   string

	RawBuildfile []byte // The raw build.yaml bytes
}

// Package is the namespace/name/version/type of a builder package
type Package struct {
	Namespace string `yaml:"namespace"` // The namespace of the package
	Name      string `yaml:"name"`      // The name of the package
	Version   string `yaml:"version"`   // The version of the package
	Type      string `yaml:"type"`      // The type of package, e.g. go, java, etc.

	Dependencies Dependencies `yaml:"dependencies"` // The set of dependencies of this package
}

// String returns a human readable string representing this package
func (p Package) String() string {
	return fmt.Sprintf("%s/%s-%s", p.Namespace, p.Name, p.Version)
}

// Dependencies is a container struct for lists of different types of dependencies
type Dependencies struct {
	Test    []Package `yaml:"test"`
	Compile []Package `yaml:"compile"`
}

// NewParsedBuildfile constructs an instance of ParsedBuildfile
func NewParsedBuildfile(buildfile *Buildfile, absoluteWorkingDir string, rawBuildfile []byte) *ParsedBuildfile {
	return &ParsedBuildfile{
		Buildfile:          *buildfile,
		AbsoluteWorkingDir: absoluteWorkingDir,
		AbsoluteBuildDir:   filepath.Join(absoluteWorkingDir, buildDir),
		RawBuildfile:       rawBuildfile,
	}
}

// ParseBuildfile parses the build file at the provided location and returns a ParsedBuildfile
func ParseBuildfile(buildfilePath string) (*ParsedBuildfile, error) {
	buildlog.Debugf("Opening buildfile %s", buildfilePath)
	buildfileFile, err := os.Open(buildfilePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening %s: %+v", buildfilePath, err)
	}

	buildlog.Debugf("Reading buildfile %s", buildfilePath)
	buildfileBytes, err := ioutil.ReadAll(buildfileFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %+v", buildfilePath, err)
	}

	buildlog.Debugf("Parsing buildfile %s", buildfilePath)
	buildfile := &Buildfile{}
	if err = yaml.Unmarshal(buildfileBytes, buildfile); err != nil {
		return nil, fmt.Errorf("Error parsing %s: %+v", buildfilePath, err)
	}

	absoluteBuildfilePath, err := filepath.Abs(buildfilePath)
	if err != nil {
		return nil, fmt.Errorf("Error determining working directory: %+v", err)
	}

	absoluteWorkingDir := filepath.Dir(absoluteBuildfilePath)
	return NewParsedBuildfile(buildfile, absoluteWorkingDir, buildfileBytes), nil
}
