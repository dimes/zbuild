// Package protobuf contains build logic for protocol buffers
package protobuf

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/copyutil"
	"github.com/dimes/zbuild/local"
	"github.com/dimes/zbuild/model"
)

const (
	protobufType = "proto"
	srcDir       = "proto"
)

// Builder contains the logic for building protocol buffers
type Builder struct {
}

// NewBuilder returns a new builder instance
func NewBuilder() *Builder {
	return &Builder{}
}

// Type returns the type of the type of this builder
func (b *Builder) Type() string {
	return protobufType
}

// Build compiles the protocol buffers to make sure the syntax is correct
func (b *Builder) Build(parsedBuildfile *model.ParsedBuildfile) error {
	buildlog.Infof("Building Protocol Buffer package %s", parsedBuildfile.Package.String())
	protoPaths, err := local.GetBuildpath(parsedBuildfile.AbsoluteWorkingDir,
		local.CompileDependencyResolver)
	if err != nil {
		return fmt.Errorf("Error getting proto path: %+v", err)
	}

	protoDir := filepath.Join(parsedBuildfile.AbsoluteWorkingDir, srcDir)
	protoFiles, err := filepath.Glob(filepath.Join(protoDir, "**", "*.proto"))
	if err != nil {
		return fmt.Errorf("Error listing proto files in %s: %+v", protoDir, err)
	}

	args := make([]string, 0)
	for _, protoPath := range protoPaths {
		args = append(args, []string{"-I", filepath.Join(protoPath, srcDir)}...)
	}

	tempDir, err := ioutil.TempDir(os.TempDir(), "protobuild")
	if err != nil {
		return fmt.Errorf("Error generating temp build directory: %+v", err)
	}
	defer os.RemoveAll(tempDir)

	args = append(args, []string{"--cpp_out", tempDir}...)
	args = append(args, protoFiles...)

	cmd := exec.Command("protoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = parsedBuildfile.AbsoluteWorkingDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error building protocol buffers in %s: %+v", protoDir, err)
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
