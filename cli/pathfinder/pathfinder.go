package main

import (
	"builder/buildlog"
	"builder/local"
	"fmt"
	"os"
	"strings"
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

	fmt.Print(strings.Join(path, string(os.PathListSeparator)))
}
