package mads

import (
	"fmt"
	"os"
	"strings"

	"github.com/arnarg/mads/pkg/mads/resource"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Module struct {
	Containers []*resource.Container `hcl:"container,block"`
	Pods       []*resource.Pod       `hcl:"pod,block"`
}

func LoadModule(path string) (*Module, error) {
	// Check if path is a file or a folder
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var body hcl.Body
	if info.IsDir() {
		b, err := bodyFromDirectory(path)
		if err != nil {
			return nil, err
		}
		body = b
	} else {
		b, diag := bodyFromFile(path)
		if diag.HasErrors() {
			return nil, diag
		}
		body = b
	}

	// Parse module from body
	module := &Module{}
	if diag := gohcl.DecodeBody(body, nil, module); diag.HasErrors() {
		return nil, diag
	}

	return module, nil
}

func bodyFromDirectory(path string) (hcl.Body, error) {
	// Get all files and directories in directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	bodies := []hcl.Body{}

	for _, entry := range entries {
		// If entry is a directory or a file not ending
		// with `.hcl` we skip it
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hcl") {
			continue
		}

		// Parse file
		body, diag := bodyFromFile(fmt.Sprintf("%s/%s", path, entry.Name()))
		if diag.HasErrors() {
			return nil, diag
		}

		bodies = append(bodies, body)
	}

	return hcl.MergeBodies(bodies), nil
}

func bodyFromFile(path string) (hcl.Body, hcl.Diagnostics) {
	parser := hclparse.NewParser()

	file, diag := parser.ParseHCLFile(path)
	if diag.HasErrors() {
		return nil, diag
	}

	return file.Body, nil
}
