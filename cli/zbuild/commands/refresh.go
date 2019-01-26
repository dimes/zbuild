package commands

import (
	"fmt"

	"github.com/dimes/zbuild/local"
)

type refresh struct{}

func (r *refresh) Describe() string {
	return "Refreshes the workspace metadata"
}

func (r *refresh) Exec(workingDir string, args ...string) error {
	workspaceDir, err := local.GetWorkspace(workingDir)
	if err != nil {
		return fmt.Errorf("Error determining workspace for %s: %+v", workingDir, err)
	}

	remoteSourceSet, err := local.GetRemoteSourceSet(workspaceDir)
	if err != nil {
		return fmt.Errorf("Error getting remote source set for %s: %+v", workspaceDir, err)
	}

	if err := local.RefreshWorkspace(workspaceDir, remoteSourceSet); err != nil {
		return fmt.Errorf("Error refreshing workspace metadata for %s: %+v", workspaceDir, err)
	}

	return nil
}
