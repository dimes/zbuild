package artifacts

import (
	"fmt"
	"regexp"

	"github.com/dimes/zbuild/model"
)

const (
	nameRegexStr        = "^[a-z0-9\\.\\-]{1,40}$"
	buildNumberRegexStr = "^[0-9]+$"
)

var (
	nameRegex        = regexp.MustCompile(nameRegexStr)
	buildNumberRegex = regexp.MustCompile(buildNumberRegexStr)
)

// IsValidName returns if a given name is valid
func IsValidName(name string) error {
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("Name %s does not match %s", name, nameRegexStr)
	}
	return nil
}

// IsValid returns true id the given artifact is valid. This only validates the data, e.g.
// ensures the names conform to the naming standards, etc.
func IsValid(artifact *model.Artifact) error {
	if !nameRegex.MatchString(artifact.Namespace) {
		return fmt.Errorf("Artifact namespace %s must match %s", artifact.Namespace, nameRegexStr)
	}

	if !nameRegex.MatchString(artifact.Name) {
		return fmt.Errorf("Artifact name %s must match %s", artifact.Name, nameRegexStr)
	}

	if !nameRegex.MatchString(artifact.Version) {
		return fmt.Errorf("Artifact version %s must match %s", artifact.Version, nameRegexStr)
	}

	if !buildNumberRegex.MatchString(artifact.BuildNumber) {
		return fmt.Errorf("Build number %s must match %s", artifact.BuildNumber, buildNumberRegexStr)
	}

	return nil
}
