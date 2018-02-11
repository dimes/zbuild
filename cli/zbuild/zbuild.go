package main

import (
	"os"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/cli/argv"
	"github.com/dimes/zbuild/cli/commands"
)

var (
	knownCommands = map[string]commands.Command{
		"build":          commands.Build,
		"init-workspace": commands.InitWorkspace,
		"publish":        commands.Publish,
		"refresh":        commands.Refresh,
	}
)

func main() {
	var verbose bool
	argSet := argv.NewArgSet()
	argSet.ExpectBool(&verbose, "v", false, "enable verbose logging")
	rest, err := argSet.Parse(os.Args[1:])
	if err != nil {
		buildlog.Fatalf("Error parsing args: %+v", err)
	}

	buildlog.SetLogLevel(buildlog.Info)
	if verbose {
		buildlog.SetLogLevel(buildlog.Debug)
	}

	if len(rest) == 0 {
		buildlog.Errorf("No command specified")
		printUsage(argSet)
		os.Exit(1)
	}

	commandName := rest[0]
	command, ok := knownCommands[commandName]
	if !ok {
		buildlog.Errorf("Unknown command %s", commandName)
		printUsage(argSet)
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		buildlog.Fatalf("Error getting working directory: %+v", err)
	}

	if err := command.Exec(workingDir, rest[1:]...); err != nil {
		buildlog.Fatalf("Error executing command %s: %+v", commandName, err)
	}
}

func printUsage(argSet *argv.ArgSet) {
	buildlog.Infof("Usage: %s [command] [options]", os.Args[0])
	buildlog.Infof("Valid commands are:")
	for commandName, command := range knownCommands {
		buildlog.Infof("\t%s\t%s", commandName, command.Describe())
	}
	buildlog.Infof("Valid options are:")
	argSet.PrintUsage()
}
