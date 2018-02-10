package main

import (
	"os"
	"strings"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/cli/argv"
	"github.com/dimes/zbuild/local"
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

	buildlog.Debugf("Getting build path for %s", file)
	path, err := local.GetBuildpath(path, dependencyResolver)
	if err != nil {
		buildlog.Fatalf("Error getting build path for %s: %+v", file, err)
	}

	buildlog.Outputf(strings.Join(path, string(os.PathListSeparator)))
}
