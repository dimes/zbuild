package main

import (
	"os"
	"strings"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/cli/argv"
	"github.com/dimes/zbuild/local"
)

func main() {
	argSet := argv.NewArgSet()
	var file string
	argSet.ExpectString(&file, "file", "", "the file to get the path for")
	argSet.Parse(os.Args[1:])

	if file == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			buildlog.Fatalf("Error getting working directory: %+v", err)
		}
		file = workingDir
	}

	// TODO: Need to determine if compile or test dependencies should be used

	buildlog.Debugf("Getting build path for %s", file)
	path, err := local.GetBuildpath(file, local.CompileDependencyResolver)
	if err != nil {
		buildlog.Fatalf("Error getting build path for %s: %+v", file, err)
	}

	buildlog.Outputf(strings.Join(path, string(os.PathListSeparator)))
}
