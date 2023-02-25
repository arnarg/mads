package delete

import (
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:        "delete",
	Aliases:     []string{"d"},
	Usage:       "Delete a single pod by name",
	Description: "This does the same as 'podman pods rm -f [POD]'",
	ArgsUsage:   "POD [POD...]",
	Action:      run,
}

func run(cCtx *cli.Context) error {
	return nil
}
