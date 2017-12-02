package commands

import (
	"bufio"
	"builder/buildlog"
	"fmt"
)

var (
	// InitWorkspace is the command that initializes a workspace on the local file system
	InitWorkspace Command = &initWorkspace{}
)

// Command is an interface for commands
type Command interface {
	Exec(workingDir string, args ...string) error
}

func readLineWithPrompt(prompt string, reader *bufio.Reader) string {
	fmt.Print(prompt)
	line, _, err := reader.ReadLine()
	if err != nil {
		buildlog.Fatalf("Error reading input: %+v", err)
	}
	return string(line)
}
