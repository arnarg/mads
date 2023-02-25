package containers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/arnarg/mads/pkg/entities"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func NewClient(c *resty.Client) *Client {
	return &Client{client: c}
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
