package pods

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func NewClient(c *resty.Client) *Client {
	return &Client{client: c}
}

// Exists checks if pod by name or ID exists.
func (p *Client) Exists(ctx context.Context, nameOrID string) (bool, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Get("/v4/libpod/pods/{id}/exists")
	if err != nil {
		return false, err
	}

	if res.StatusCode() == 204 {
		return true, nil
	}

	return false, nil
}

// Inspect returns info about pod.
func (p *Client) Inspect(ctx context.Context, nameOrID string) (*PodInspectResponse, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Get("/v4/libpod/pods/{id}/json")
	if err != nil {
		return nil, err
	}

	// Parse JSON
	pod := &PodInspectResponse{}
	err = json.Unmarshal(res.Body(), pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// Create creates a new pod.
func (p *Client) Create(ctx context.Context, pod *PodCreateRequest) error {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetBody(pod).
		Post("/v4/libpod/pods/create")
	if err != nil {
		return err
	}

	if res.StatusCode() != 201 {
		return fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	return nil
}
