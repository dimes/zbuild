package protobuf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dimes/zbuild/local"
	"github.com/dimes/zbuild/model"
	yaml "gopkg.in/yaml.v2"
)

const (
	protogenType = "protogen"
)

var (
	protoOpts = map[string]*protoOpt{
		"go": {
			flag:      "--go_out",
			outputDir: "src",
		},
	}
)

type protoOpt struct {
	flag      string
	outputDir string
}

// ProtogenBuildfile contains protogen specific build options
type ProtogenBuildfile struct {
	Protogen struct {
		Lang   string        `yaml:"lang,omitempty"`
		Source model.Package `yaml:"source,omitempty"`
	} `yaml:"protogen,omitempty"`
}

// Protogen is a builder that generates source code
type Protogen struct {
}

// NewProtogen returns a new protogen builder
func NewProtogen() *Protogen {
	return &Protogen{}
}

// Type returns the type of the type of this builder
func (p *Protogen) Type() string {
	return protogenType
}

// Build compiles the protocol buffers to make sure the syntax is correct
func (p *Protogen) Build(workspace string, parsedBuildfile *model.ParsedBuildfile) error {
	protogenBuildfile := &ProtogenBuildfile{}
	if err := yaml.Unmarshal(parsedBuildfile.RawBuildfile, protogenBuildfile); err != nil {
		return fmt.Errorf("Error parsing protogen buildfile: %+v", err)
	}

	lang := protogenBuildfile.Protogen.Lang
	if lang == "" {
		return fmt.Errorf("The `lang` option must be set")
	}

	langOpt, ok := protoOpts[lang]
	if !ok {
		return fmt.Errorf("Unknown lang %s", lang)
	}

	artifactDir, err := local.GetArtifactLocation(parsedBuildfile.AbsoluteWorkingDir,
		protogenBuildfile.Protogen.Source)
	if err != nil {
		return fmt.Errorf("Error getting source location: %+v", err)
	}

	protoDir := filepath.Join(artifactDir, srcDir)
	protoFiles, err := filepath.Glob(filepath.Join(protoDir, "**", "*.proto"))
	if err != nil {
		return fmt.Errorf("Error listing proto files in %s: %+v", protoDir, err)
	}

	protoPaths, err := local.GetBuildpath(workspace, protogenBuildfile.Protogen.Source,
		local.CompileDependencyResolver)
	if err != nil {
		return fmt.Errorf("Error getting proto path: %+v", err)
	}

	outputDir := filepath.Join(parsedBuildfile.AbsoluteBuildDir, langOpt.outputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("Error making output directory %s: %+v", outputDir, err)
	}

	err = runProtoc(
		parsedBuildfile.AbsoluteWorkingDir,
		protoPaths,
		protoFiles,
		langOpt.flag, outputDir)
	if err != nil {
		return err
	}

	return nil
}
