package main

import (
	"builder/buildlog"
	"builder/cli/argv"
	"builder/cli/commands"
	"os"
)

var (
	knownCommands = map[string]commands.Command{
		"init-workspace": commands.InitWorkspace,
	}
)

func main() {
	var verbose bool
	argSet := argv.NewArgSet()
	argSet.ExpectBool(&verbose, "v", false, "if set, verbose logging will be enabled")
	rest, err := argSet.Parse(os.Args[1:])
	if err != nil {
		buildlog.Fatalf("Error parsing args: %+v", err)
	}

	buildlog.SetLogLevel(buildlog.Info)
	if verbose {
		buildlog.SetLogLevel(buildlog.Debug)
	}

	if len(rest) == 0 {
		buildlog.Fatalf("No command specified")
	}

	commandName := rest[0]
	command, ok := knownCommands[commandName]
	if !ok {
		buildlog.Fatalf("Unknown command %s", commandName)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		buildlog.Fatalf("Error getting working directory: %+v", err)
	}

	if err := command.Exec(workingDir, rest[1:]...); err != nil {
		buildlog.Fatalf("Error executing command %s: %+v", commandName, err)
	}
}
