package pods

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

var (
	ErrPodAlreadyStarted = errors.New("pod already started")
)

type Client struct {
	client *resty.Client
}

func NewClient(c *resty.Client) *Client {
	return &Client{client: c}
}

// Exists checks if pod by name or ID exists.
func (p *Client) Exists(ctx context.Context, nameOrID string) (bool, string, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Get("/v4/libpod/pods/{id}/exists")
	if err != nil {
		return false, "", err
	}

	if res.StatusCode() != 204 {
		return false, "", nil
	}

	// Get pod ID
	info, err := p.Inspect(ctx, nameOrID)
	if err != nil {
		return false, "", err
	}

	return true, info.Id, nil
}

// Inspect returns info about pod.
func (p *Client) Inspect(ctx context.Context, nameOrID string) (*PodInfo, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Get("/v4/libpod/pods/{id}/json")
	if err != nil {
		return nil, err
	}

	// Parse JSON
	pod := &PodInfo{}
	err = json.Unmarshal(res.Body(), pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// Create creates a new pod.
func (p *Client) Create(ctx context.Context, pod *PodCreateRequest) (string, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetBody(pod).
		Post("/v4/libpod/pods/create")
	if err != nil {
		return "", err
	}

	if res.StatusCode() != 201 {
		return "", fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	// Get pod ID
	info, err := p.Inspect(ctx, pod.Name)
	if err != nil {
		return "", err
	}

	return info.Id, nil
}

// Delete deletes a pod
func (p *Client) Delete(ctx context.Context, nameOrID string, force bool) error {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetQueryParam("force", strconv.FormatBool(force)).
		SetPathParam("id", nameOrID).
		Delete("/v4/libpod/pods/{id}")
	if err != nil {
		return err
	}

	if res.StatusCode() != 200 {
		return fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	return nil
}

// Start starts a pod.
func (p *Client) Start(ctx context.Context, nameOrID string) error {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Post("/v4/libpod/pods/{id}/start")
	if err != nil {
		return err
	}

	if res.StatusCode() == 304 {
		return ErrPodAlreadyStarted
	} else if res.StatusCode() != 200 {
		return fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	return nil
}
