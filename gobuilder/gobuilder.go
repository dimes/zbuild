// Package gobuilder contains all logic related to building go code
package gobuilder

import (
	"builder/buildlog"
	"builder/copyutil"
	"builder/local"
	"builder/model"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	goBuilderType = "go"
	srcDir        = "src"
	envFormat     = "%s=%s"
)

// GoBuilder contains most of the logic for building Go code
type GoBuilder struct {
}

// NewGoBuilder returns a new instance of the go builder
func NewGoBuilder() *GoBuilder {
	return &GoBuilder{}
}

// Type returns the type this builder should be registered under
func (g *GoBuilder) Type() string {
	return goBuilderType
}

// Build implements the Builder's Build method.
//
// Go builds consist of compiling all the code (to make sure it builds)
// and then copying the source files to the build directory.
func (g *GoBuilder) Build(parsedBuildfile *model.ParsedBuildfile) error {
	buildlog.Infof("Building Go package %s", parsedBuildfile.Package.String())
	env, err := generateEnvironment(parsedBuildfile)
	if err != nil {
		return fmt.Errorf("Error generating build environment: %+v", err)
	}

	cmd := exec.Command("go", "build", filepath.Join(".", "..."))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = parsedBuildfile.AbsoluteWorkingDir
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error building %s: %+v", parsedBuildfile.Package.String(), err)
	}

	buildlog.Infof("Copying source files to build directory %s", parsedBuildfile.AbsoluteBuildDir)
	absoluteSrcOutput := filepath.Join(parsedBuildfile.AbsoluteBuildDir, srcDir)
	absoluteSrcInput := filepath.Join(parsedBuildfile.AbsoluteWorkingDir, srcDir)
	buildlog.Debugf("Beginning to copy input source %s to %s", absoluteSrcInput, absoluteSrcOutput)

	if err := copyutil.Copy(absoluteSrcInput, absoluteSrcOutput); err != nil {
		return fmt.Errorf("Error copying source file to %s: %+v", absoluteSrcOutput, err)
	}

	return nil
}

func generateEnvironment(parsedBuildfile *model.ParsedBuildfile) ([]string, error) {
	gopath, err := local.GetBuildpath(parsedBuildfile.AbsoluteWorkingDir, local.CompileDependencyResolver)
	if err != nil {
		return nil, fmt.Errorf("Error getting GOPATH: %+v", err)
	}

	env := make([]string, 0)
	env = append(env, fmt.Sprintf(envFormat, "GOPATH", strings.Join(gopath, string(os.PathListSeparator))))
	return env, nil
}
