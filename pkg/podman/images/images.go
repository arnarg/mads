package images

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func NewClient(c *resty.Client) *Client {
	return &Client{client: c}
}

// Exists check if image by name or ID exists.
func (p *Client) Exists(ctx context.Context, nameOrID string) (bool, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Get("/v4/libpod/images/{id}/exists")
	if err != nil {
		return false, err
	}

	if res.StatusCode() == 204 {
		return true, nil
	}

	return false, nil
}

// Inspect returns detailed info about an image.
func (p *Client) Inspect(ctx context.Context, nameOrID string) (*ImageInfo, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Get("/v4/libpod/images/{id}/json")
	if err != nil {
		return nil, err
	}

	// Parse JSON
	img := &ImageInfo{}
	err = json.Unmarshal(res.Body(), img)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// Load loads image from reader and imports into podman.
func (p *Client) Load(ctx context.Context, image io.Reader) (*ImageInfo, error) {
	res, err := p.client.R().
		ForceContentType("application/x-tar").
		SetBody(image).
		Post("/v4/libpod/images/load")
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	ilr := &ImageLoadResponse{}
	err = json.Unmarshal(res.Body(), ilr)
	if err != nil {
		return nil, err
	}

	if len(ilr.Names) < 1 {
		return nil, fmt.Errorf("load failed")
	}

	return p.Inspect(ctx, ilr.Names[0])
}

// Pull pulls an image.
func (p *Client) Pull(ctx context.Context, image string, opts *PullOptions) (*ImageInfo, error) {
	if opts == nil {
		opts = &PullOptions{Policy: PullPolicyAlways}
	}

	// Make request
	res, err := p.client.R().
		SetQueryParams(map[string]string{
			"reference": image,
			"policy":    opts.Policy,
		}).
		Post("/v4/libpod/images/pull")
	if err != nil {
		return nil, err
	}

	// Split each line of json
	lines := bytes.SplitAfter(res.Body(), []byte{'\n'})
	last := lines[len(lines)-2]

	// Parse JSON
	pullData := &ImagePullResponse{}
	err = json.Unmarshal(last, pullData)

	if pullData.Error != "" {
		return nil, fmt.Errorf(pullData.Error)
	}

	return p.Inspect(ctx, pullData.Id)
}
