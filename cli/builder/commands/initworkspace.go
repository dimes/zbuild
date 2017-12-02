package commands

import (
	"builder/local"
	"fmt"
)

type initWorkspace struct {
}

func (i *initWorkspace) Exec(workingDir string, args ...string) error {
	if workspaceDir, err := local.GetWorkspace(workingDir); err == nil {
		return fmt.Errorf("Workspace already exists at %s", workspaceDir)
	} else if err != local.ErrWorkspaceNotFound {
		return fmt.Errorf("Error validating no existing workspace: %+v", err)
	}

	// reader := bufio.NewReader(os.Stdin)

	return nil
}
