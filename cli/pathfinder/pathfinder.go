package main

import (
	"os"
	"strings"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/local"
)

func main() {
	workingDir, err := os.Getwd()
	if err != nil {
		buildlog.Fatalf("Error getting working directory: %+v", err)
	}

	path, err := local.GetBuildpath(workingDir, local.CompileDependencyResolver)
	if err != nil {
		buildlog.Fatalf("Error getting build path for %s: %+v", workingDir, err)
	}

	buildlog.Outputf(strings.Join(path, string(os.PathListSeparator)))
}
