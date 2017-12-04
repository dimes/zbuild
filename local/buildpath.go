package local

import (
	"builder/artifacts"
	"builder/buildlog"
	"builder/model"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// GetBuildpath returns a path to all packages required for the build
func GetBuildpath(path string, resolver DependencyResolver) ([]string, error) {
	// Calculate the package containing the path
	workspace, err := GetWorkspace(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting workspace for %s: %+v", path, err)
	}

	relativePath, err := filepath.Rel(workspace, path)
	if err != nil {
		return nil, fmt.Errorf("Error getting path %s relative to %s: %+v", path, workspace, err)
	}

	packageLocation := filepath.Join(workspace, strings.Split(relativePath, string(os.PathSeparator))[0])
	parsedBuildfile, err := model.ParseBuildfile(filepath.Join(packageLocation, model.BuildfileName))
	if err != nil {
		return nil, fmt.Errorf("Error parsing build file for package %s: %+v", packageLocation, err)
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
		buildlog.Fatalf("Error getting remote manager: %+v", err)
	}

	paths := make([]string, 0)
	seenPackages := make(map[string]bool)
	stack := []*stackEntry{{target: parsedBuildfile.Package}}
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

		artifactLocation := ""
		artifact, err := overrideSourceSet.GetArtifact(target.Namespace, target.Name, target.Version)
		if err == artifacts.ErrArtifactNotFound {
			artifact, err = localSourceSet.GetArtifact(target.Namespace, target.Name, target.Version)
			if err != nil {
				return nil, fmt.Errorf("Error getting artifact for %s: %+v", target.String(), err)
			}

			artifactLocation = localArtifactCacheDir(workspace, artifact)
			if _, err = os.Stat(artifactLocation); err != nil {
				buildlog.Debugf("Downloading %s", artifact.String())
				if err = artifacts.Transfer(upstreamManager, localManager, artifact); err != nil {
					return nil, fmt.Errorf("Error downloading artifact %s: %+v", artifact.String(), err)
				}
			}
		} else if err != nil {
			return nil, fmt.Errorf("Error getting artifact from overide source set: %+v", err)
		} else {
			artifactLocation, err = overrideSourceSet.getLocationForArtifact(target.Namespace, target.Name,
				target.Version)
			if err != nil {
				return nil, fmt.Errorf("Error getting artifact location for %s: %+v", artifact.String(), err)
			}
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
