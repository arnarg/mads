package main

import (
	"log"
	"os"

	"github.com/arnarg/mads/cmd/mads/agent"
	"github.com/arnarg/mads/cmd/mads/apply"
	"github.com/arnarg/mads/cmd/mads/delete"
	"github.com/urfave/cli/v2"
)

var (
	version = "unknown"
)

func main() {
	app := &cli.App{
		Name:        "mads",
		Version:     version,
		Description: "Run pods in podman with consul services from a declarative definition.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "socket",
				Aliases: []string{"s"},
				EnvVars: []string{"MADS_PODMAN_SOCKET"},
				Value:   "$XDG_RUNTIME_DIR/podman/podman.sock",
			},
		},
		Before: func(cCtx *cli.Context) error {
			// Expand env variable in socket flag
			rsocket := os.ExpandEnv(cCtx.String("socket"))

			// Update value
			return cCtx.Set("socket", rsocket)

		},
		Commands: cli.Commands{
			apply.Command,
			delete.Command,
			agent.Command,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
