package delete

import (
	"context"

	"github.com/arnarg/mads/pkg/orchestrator"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:        "delete",
	Aliases:     []string{"d"},
	Usage:       "Delete a single pod by name",
	Description: "Deletes pods by name and cleans up their consul services.",
	ArgsUsage:   "POD [POD...]",
	Action:      run,
}

func run(cCtx *cli.Context) error {
	// Get podman socket path
	socket := cCtx.String("socket")

	// Get list of pod names to delete
	podNames := cCtx.Args().Slice()

	// Create orchestrator instance
	orch, err := orchestrator.NewOrchestrator(&orchestrator.Config{
		PodmanSocketPath: socket,
	})
	if err != nil {
		return err
	}

	// Iterate over pods and delete them
	for _, n := range podNames {
		// Delete pod
		err := orch.Delete(context.Background(), n)
		if err != nil {
			return err
		}
	}

	return nil
}
