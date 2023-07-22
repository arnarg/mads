package containers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/arnarg/mads/pkg/entities"
	"github.com/go-resty/resty/v2"
)

var (
	ErrContainerAlreadyStarted = errors.New("container already started")
)

type Client struct {
	client *resty.Client
}

func NewClient(c *resty.Client) *Client {
	return &Client{client: c}
}

// Exists checks if container by name or ID exists.
func (p *Client) Exists(ctx context.Context, nameOrID string) (bool, string, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Get("/v4/libpod/containers/{id}/exists")
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

// Inspect returns info about a container.
func (p *Client) Inspect(ctx context.Context, nameOrID string) (*ContainerInfo, error) {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Get("/v4/libpod/containers/{id}/json")
	if err != nil {
		return nil, err
	}

	// Parse JSON
	cnt := &ContainerInfo{}
	err = json.Unmarshal(res.Body(), cnt)
	if err != nil {
		return nil, err
	}

	return cnt, nil
}

// Create creates a new container.
func (c *Client) Create(ctx context.Context, ctr *ContainerCreateRequest) error {
	res, err := c.client.R().
		ForceContentType("application/json").
		SetBody(ctr).
		Post("/v4/libpod/containers/create")
	if err != nil {
		return err
	}

	// Handle error message
	if res.StatusCode() >= 400 {
		e := &entities.PodmanAPIError{}
		err := json.Unmarshal(res.Body(), e)
		if err != nil {
			return fmt.Errorf("could not parse error message")
		}
		return e
	}

	if res.StatusCode() != 201 {
		return fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	return nil
}

// Start starts a container
func (p *Client) Start(ctx context.Context, nameOrID string) error {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetPathParam("id", nameOrID).
		Post("/v4/libpod/containers/{id}/start")
	if err != nil {
		return err
	}

	if res.StatusCode() == 304 {
		return ErrContainerAlreadyStarted
	} else if res.StatusCode() != 204 {
		return fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	return nil
}

// Delete deletes a container
func (p *Client) Delete(ctx context.Context, nameOrID string, force bool) error {
	res, err := p.client.R().
		ForceContentType("application/json").
		SetQueryParam("force", strconv.FormatBool(force)).
		SetPathParam("id", nameOrID).
		Delete("/v4/libpod/containers/{id}")
	if err != nil {
		return err
	}

	if res.StatusCode() != 200 {
		return fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	return nil
}

func (c *Client) Copy(ctx context.Context, nameOrID string, w io.Reader) error {
	res, err := c.client.R().
		ForceContentType("application/x-tar").
		SetQueryParam("path", "/").
		SetBody(w).
		SetPathParam("id", nameOrID).
		Put("/v4/libpod/containers/{id}/archive")
	if err != nil {
		return err
	}

	// Handle error message
	if res.StatusCode() >= 400 {
		e := &entities.PodmanAPIError{}
		err := json.Unmarshal(res.Body(), e)
		if err != nil {
			return fmt.Errorf("could not parse error message")
		}
		return e
	}

	if res.StatusCode() != 200 {
		return fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	return nil
}

// CopyFile copies an archive and extracts it into a container
func (c *Client) CopyFile(ctx context.Context, nameOrID string, p string) error {
	res, err := c.client.R().
		ForceContentType("application/json").
		SetQueryParam("path", "/").
		SetBody(p).
		SetPathParam("id", nameOrID).
		Put("/v4/libpod/containers/{id}/archive")
	if err != nil {
		return err
	}

	// Handle error message
	if res.StatusCode() >= 400 {
		e := &entities.PodmanAPIError{}
		err := json.Unmarshal(res.Body(), e)
		if err != nil {
			return fmt.Errorf("could not parse error message")
		}
		return e
	}

	if res.StatusCode() != 200 {
		return fmt.Errorf("unknown status code %d", res.StatusCode())
	}

	return nil
}
