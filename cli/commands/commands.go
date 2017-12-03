package commands

import (
	"builder/buildlog"
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

var (
	// InitWorkspace is the command that initializes a workspace on the local file system
	InitWorkspace Command = &initWorkspace{}

	// Build is the command that executes a build
	Build Command = &build{}
)

// Command is an interface for commands
type Command interface {
	Exec(workingDir string, args ...string) error
}

func readLineWithPrompt(label string, validate promptui.ValidateFunc) string {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		buildlog.Fatalf("Error getting option for label %s: %+v", label, err)
	}

	return result
}

func getYnConfirmation(label string) bool {
	result := readLineWithPrompt(fmt.Sprintf("%s (y/n)", label), func(input string) error {
		input = strings.ToLower(input)
		if input == "y" || input == "n" {
			return nil
		}
		return errors.New("Input must be 'y' or 'n'")
	})

	result = strings.ToLower(result)
	return result == "y"
}
