package main

import (
	"builder/buildlog"
	"builder/cli/argv"
	"builder/cli/commands"
	"fmt"
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
	fmt.Printf("Usage: %s [command] [options]\n", os.Args[0])
	fmt.Println("Valid commands are:")
	for commandName, command := range knownCommands {
		fmt.Printf("\t%s\t%s\n", commandName, command.Describe())
	}
	fmt.Println("\nValid options are:")
	argSet.PrintUsage()
}
