package local

import (
	"fmt"
	"os"
	"strings"

	"github.com/dimes/zbuild/artifacts"
	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/model"
)

var (
	// CompileDependencyResolver resolves compile-time dependencies
	CompileDependencyResolver DependencyResolver = &compileResolver{}

	// TestDependencyResolver resolves test-dependencies
	TestDependencyResolver DependencyResolver = &testResolver{compileResolver: &compileResolver{}}
)

// DependencyResolver resolves different types of dependencies, e.g. test, compile, runtime, etc.
type DependencyResolver interface {
	GetDependencies(target model.Package) []model.Package
}

type compileResolver struct {
}

func (c *compileResolver) GetDependencies(target model.Package) []model.Package {
	return target.Dependencies.Compile
}

type testResolver struct {
	compileResolver *compileResolver
}

func (t *testResolver) GetDependencies(target model.Package) []model.Package {
	return append(t.compileResolver.GetDependencies(target), target.Dependencies.Test...)
}

type buildpathGenerator struct {
	workspace         string
	localSourceSet    artifacts.SourceSet
	overrideSourceSet *overrideSourceSet
	localManager      artifacts.Manager
	upstreamManager   artifacts.Manager
}

func newBuildpathGenerator(path string) (*buildpathGenerator, error) {
	workspace, err := GetWorkspace(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace for %s: %+v", path, err)
	}

	localSourceSet, err := NewLocalSourceSet(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error creating local source set: %+v", err)
	}

	overrideSourceSet, err := newOverrideSourceSet(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error creating override source set: %+v", err)
	}

	localManager, err := NewLocalManager(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error creating local manager: %+v", err)
	}

	upstreamManager, err := GetRemoteManager(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error getting remote manager: %+v", err)
	}

	return &buildpathGenerator{
		workspace:         workspace,
		localSourceSet:    localSourceSet,
		overrideSourceSet: overrideSourceSet,
		localManager:      localManager,
		upstreamManager:   upstreamManager,
	}, nil
}

// getArtifact returns the artifact for the given package, as well as its location in the local FS
func (b *buildpathGenerator) getArtifact(target model.Package) (*model.Artifact, string, error) {
	artifactLocation := ""
	artifact, err := b.overrideSourceSet.GetArtifact(target.Namespace, target.Name, target.Version)
	if err == artifacts.ErrArtifactNotFound {
		artifact, err = b.localSourceSet.GetArtifact(target.Namespace, target.Name, target.Version)
		if err != nil {
			return nil, "", fmt.Errorf("Error getting artifact for %s: %+v", target.String(), err)
		}

		artifactLocation = localArtifactCacheDir(b.workspace, artifact)
		if _, err = os.Stat(artifactLocation); err != nil {
			buildlog.Debugf("Downloading %s", artifact.String())
			if err = artifacts.Transfer(b.upstreamManager, b.localManager, artifact); err != nil {
				return nil, "", fmt.Errorf("Error downloading artifact %s: %+v", artifact.String(), err)
			}
		}
	} else if err != nil {
		return nil, "", fmt.Errorf("Error getting artifact from overide source set: %+v", err)
	} else {
		artifactLocation, err = b.overrideSourceSet.getLocationForArtifact(
			target.Namespace,
			target.Name,
			target.Version)
		if err != nil {
			return nil, "",
				fmt.Errorf("Error getting artifact location for %s: %+v", artifact.String(), err)
		}
	}

	return artifact, artifactLocation, nil
}

// GetArtifactLocation gets the artifact for the given package. It uses the path to determine the
// workspace
func GetArtifactLocation(path string, target model.Package) (string, error) {
	buildpathGenerator, err := newBuildpathGenerator(path)
	if err != nil {
		return "", fmt.Errorf("Error getting buildpath generator for %s: %+v", path, err)
	}

	_, artifactLocation, err := buildpathGenerator.getArtifact(target)
	if err != nil {
		return "", fmt.Errorf("Error getting artifact for %+v: %+v", target, err)
	}

	return artifactLocation, nil
}

// GetBuildpath returns a path to all packages required for the build
func GetBuildpath(workspace string, target model.Package, resolver DependencyResolver) ([]string, error) {
	buildpathGenerator, err := newBuildpathGenerator(workspace)
	if err != nil {
		return nil, fmt.Errorf("Error getting buildpath generator for %s: %+v", workspace, err)
	}

	paths := make([]string, 0)
	seenPackages := make(map[string]bool)
	stack := []*stackEntry{{target: target}}
	for len(stack) > 0 {
		entry := stack[len(stack)-1]
		target := entry.target
		targetKey := packageToMapKey(target)
		if entry.visited {
			delete(seenPackages, targetKey)
			stack = stack[:len(stack)-1]
			continue
		}

		if seenPackages[targetKey] {
			cycle := make([]string, 0)
			for _, entry := range stack {
				cycle = append(cycle, packageToMapKey(entry.target))
			}
			return nil, fmt.Errorf("Dependency cycle detected: %s", strings.Join(cycle, " -> "))
		}

		seenPackages[targetKey] = true
		entry.visited = true

		artifact, artifactLocation, err := buildpathGenerator.getArtifact(target)
		if err != nil {
			return nil, fmt.Errorf("Error getting artifact for %+v: %+v", target, err)
		}

		paths = append(paths, artifactLocation)
		dependencies := resolver.GetDependencies(artifact.Package)
		for _, dependency := range dependencies {
			stack = append(stack, &stackEntry{target: dependency})
		}
	}

	return paths, nil
}

func packageToMapKey(target model.Package) string {
	return packageInfoToMapKey(target.Namespace, target.Name, target.Version)
}

func packageInfoToMapKey(namespace, name, version string) string {
	return fmt.Sprintf("%s/%s/%s", namespace, name, version)
}

type stackEntry struct {
	target  model.Package
	visited bool
}
