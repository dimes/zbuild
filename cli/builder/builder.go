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
	rest, _ := argSet.Parse(os.Args[1:])

	if len(rest) == 0 {
		buildlog.Fatalf("No command specified")
	}

	commandName := rest[0]
	command, ok := knownCommands[commandName]
	if !ok {
		buildlog.Fatalf("Unknown command %s", commandName)
	}

	if err := command.Exec("abcd", rest[1:]...); err != nil {
		buildlog.Fatalf("Error executing command %s: %+v", commandName, err)
	}
}
