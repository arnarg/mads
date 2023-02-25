package apply

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/arnarg/mads/pkg/entities"
	"github.com/arnarg/mads/pkg/orchestrator"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var Command = &cli.Command{
	Name:        "apply",
	Aliases:     []string{"a"},
	Usage:       "Apply a single pod definition file",
	Description: "Reads provided file and applies it to podman",
	ArgsUsage:   "FILE [FILE...]",
	Action:      run,
}

func run(cCtx *cli.Context) error {
	// Get podman socket path
	socket := cCtx.String("socket")

	// Get list of pod definition paths
	paths := cCtx.Args().Slice()

	// Parse all pod definition files
	pods := []*entities.Pod{}
	for _, fpath := range paths {
		// Resolve path
		rpath, err := filepath.Abs(fpath)
		if err != nil {
			return fmt.Errorf("could not resolve path '%s': %s", fpath, err)
		}

		// Read file
		def, err := ioutil.ReadFile(rpath)
		if err != nil {
			return fmt.Errorf("could not read file '%s': %s", fpath, err)
		}

		// Parse file contents
		pod := &entities.Pod{}
		err = yaml.Unmarshal(def, pod)
		if err != nil {
			return fmt.Errorf("could not parse yaml file '%s': %s", fpath, err)
		}

		// Add to slice of pods
		pods = append(pods, pod)
	}

	// Create an orchestrator instance
	orch, err := orchestrator.NewOrchestrator(&orchestrator.Config{
		PodmanSocketPath: socket,
	})
	if err != nil {
		return err
	}

	// Apply all pods
	for _, pod := range pods {
		err := orch.Apply(context.Background(), pod)
		if err != nil {
			return fmt.Errorf("could not apply pod '%s': %s", pod.Name, err)
		}
	}

	return nil
}
