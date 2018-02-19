package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/cli/argv"
	"github.com/dimes/zbuild/local"
	"github.com/dimes/zbuild/model"
)

const (
	compileResolver = "compile"
	testResolver    = "test"
)

func main() {
	argSet := argv.NewArgSet()
	var path string
	var resolver string
	argSet.ExpectString(&path, "path", "", "the file to get the path for")
	argSet.ExpectString(&resolver, "resolver", compileResolver, "the type of dependency resolver to use")
	argSet.Parse(os.Args[1:])

	if path == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			buildlog.Fatalf("Error getting working directory: %+v", err)
		}
		path = workingDir
	}

	var dependencyResolver local.DependencyResolver
	switch resolver {
	case compileResolver:
		dependencyResolver = local.CompileDependencyResolver
	case testResolver:
		dependencyResolver = local.TestDependencyResolver
	default:
		buildlog.Fatalf("Could not find dependency resolver of type %s", resolver)
	}

	workspace, err := local.GetWorkspace(path)
	if err != nil {
		buildlog.Fatalf("Error getting workspace for %s: %+v", path, err)
	}

	relativePath, err := filepath.Rel(workspace, path)
	if err != nil {
		buildlog.Fatalf("Error getting path %s relative to %s: %+v", path, workspace, err)
	}

	packageLocation := filepath.Join(workspace, strings.Split(relativePath, string(os.PathSeparator))[0])
	parsedBuildfile, err := model.ParseBuildfile(filepath.Join(packageLocation, model.BuildfileName))
	if err != nil {
		buildlog.Fatalf("Error parsing build file for package %s: %+v", packageLocation, err)
	}

	buildlog.Debugf("Getting build path for %s", path)
	buildpath, err := local.GetBuildpath(workspace, parsedBuildfile.Package, dependencyResolver)
	if err != nil {
		buildlog.Fatalf("Error getting build path for %s: %+v", path, err)
	}

	buildlog.Outputf(strings.Join(buildpath, string(os.PathListSeparator)))
}
