package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/arnarg/mads/pkg/orchestrator"
	"github.com/arnarg/mads/pkg/watcher"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:        "agent",
	Aliases:     []string{"ag"},
	Usage:       "Run mads agent",
	Description: "Watches a directory for pod definition files and applies them",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "watch-dir",
			Aliases: []string{"w"},
			EnvVars: []string{"MADS_WATCH_DIR"},
		},
	},
	Before: before,
	Action: run,
}

func before(cCtx *cli.Context) error {
	w := cCtx.String("watch-dir")
	if w == "" {
		return fmt.Errorf("watch-dir must be specified")
	}

	// Get absolute path
	p, err := filepath.Abs(w)
	if err != nil {
		return err
	}

	// Check that watch directory exists
	stat, err := os.Stat(p)
	if err != nil {
		return err
	}

	// Check that it is a directory
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", p)
	}

	// Update flag value to absolute path
	cCtx.Set("watch-dir", p)

	return nil
}

func run(cCtx *cli.Context) error {
	// Get podman socket path
	socket := cCtx.String("socket")

	// Get watch-dir path
	watchDir := cCtx.String("watch-dir")

	// Create orchestrator instance
	orch, err := orchestrator.NewOrchestrator(&orchestrator.Config{
		PodmanSocketPath: socket,
	})
	if err != nil {
		return err
	}

	// Create a file watcher
	w := watcher.NewFileWatcher(watchDir)

	// Create an app context
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run watcher
	wg := sync.WaitGroup{}
	errCh := make(chan error)
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Run watcher
		err := w.Run(appCtx)
		if err != nil {
			errCh <- err
		}
	}()

	// Catch sigint
	intChan := make(chan os.Signal, 10)
	signal.Notify(intChan, os.Interrupt, syscall.SIGTERM) // Stop running

	// Wait for events
	for {
		select {
		// File event
		case ev, ok := <-w.PodFileEvents():
			if !ok {
				return fmt.Errorf("pod file event channel unexpectedly closed")
			}

			switch ev.Type {
			// Pod should be applied
			case watcher.TypeApply:
				log.Printf("applying pod '%s'", ev.Pod.Name)

				err := orch.Apply(appCtx, ev.Pod)
				if err != nil {
					log.Printf("could not apply pod '%s': %s", ev.Pod.Name, err)
				}

			// Pod should be deleted
			case watcher.TypeDelete:
				log.Printf("deleting pod '%s'", ev.Name)

				err := orch.Delete(appCtx, ev.Name)
				if err != nil {
					log.Printf("could not delete pod '%s': %s", ev.Name, err)
				}
			}

		// Get error from watcher
		case err := <-errCh:
			return err

		// sigint caught
		case <-intChan:
			cancel()
			wg.Wait()
			return nil
		}
	}
}
