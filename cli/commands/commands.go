package commands

import (
	"builder/buildlog"

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
	Describe() string
	Exec(workingDir string, args ...string) error
}

func readLineWithPrompt(label string, validate promptui.ValidateFunc, defaultVal string) string {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
		Default:  defaultVal,
	}

	result, err := prompt.Run()
	if err != nil {
		buildlog.Fatalf("Error getting option for label %s: %+v", label, err)
	}

	return result
}

func getYnConfirmation() (bool, error) {
	prompt := promptui.Select{
		Label: "Confirm?",
		Items: []string{"Yes", "No"},
	}

	selectedIndex, _, err := prompt.Run()
	if err != nil {
		return false, err
	}

	return selectedIndex == 0, nil
}
