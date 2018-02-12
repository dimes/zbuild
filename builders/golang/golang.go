// Package golang contains all logic related to building go code
package golang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/copyutil"
	"github.com/dimes/zbuild/local"
	"github.com/dimes/zbuild/model"
	yaml "gopkg.in/yaml.v2"
)

const (
	goBuilderType = "go"
	srcDir        = "src"
	binDir        = "bin"
	envFormat     = "%s=%s"
)

// Buildfile contains Go specific build options
type Buildfile struct {
	Go struct {
		Targets []string `yaml:"targets,omitempty"`
	} `yaml:"go,omitempty"`
}

// Builder contains most of the logic for building Go code
type Builder struct {
}

// NewBuilder returns a new instance of the go builder
func NewBuilder() *Builder {
	return &Builder{}
}

// Type returns the type this builder should be registered under
func (g *Builder) Type() string {
	return goBuilderType
}

// Build implements the Builder's Build method.
//
// Go builds consist of compiling all the code (to make sure it builds)
// and then copying the source files to the build directory.
func (g *Builder) Build(parsedBuildfile *model.ParsedBuildfile) error {
	buildlog.Infof("Building Go package %s", parsedBuildfile.Package.String())
	env, err := generateEnvironment(parsedBuildfile)
	if err != nil {
		return fmt.Errorf("Error generating build environment: %+v", err)
	}

	goBuildfile := &Buildfile{}
	if err = yaml.Unmarshal(parsedBuildfile.RawBuildfile, goBuildfile); err != nil {
		return fmt.Errorf("Error parsing go buildfile: %+v", err)
	}

	cmd := exec.Command("go", "build", filepath.Join(".", "..."))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = parsedBuildfile.AbsoluteWorkingDir
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error building %s: %+v", parsedBuildfile.Package.String(), err)
	}

	absoluteBinDir := filepath.Join(parsedBuildfile.AbsoluteBuildDir, binDir)
	if len(goBuildfile.Go.Targets) > 0 {
		os.Mkdir(absoluteBinDir, os.ModePerm)
	}

	for _, target := range goBuildfile.Go.Targets {
		buildlog.Infof("Building target %s", target)
		targetName := filepath.Base(filepath.Dir(target))
		if targetName == "" || targetName == string(os.PathSeparator) {
			return fmt.Errorf("Could not determine executable name for target %s", target)
		}

		cmd := exec.Command("go", "build", "-o", filepath.Join(absoluteBinDir, targetName), target)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = parsedBuildfile.AbsoluteWorkingDir
		cmd.Env = env
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Error building target %s: %+v", target, err)
		}
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
