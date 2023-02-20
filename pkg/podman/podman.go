package podman

import (
	"net"
	"net/http"

	"github.com/arnarg/mads/pkg/podman/images"
	"github.com/arnarg/mads/pkg/podman/pods"
	"github.com/go-resty/resty/v2"
)

type Config struct {
	SocketPath string
}

type Client struct {
	client *resty.Client
}

func NewClient(cfg *Config) *Client {
	transport := &http.Transport{
		Dial: func(_, _ string) (net.Conn, error) {
			return net.Dial("unix", cfg.SocketPath)
		},
	}

	client := resty.New().
		SetTransport(transport).
		SetScheme("http").
		SetContentLength(true).
		SetHostURL("d")

	return &Client{
		client: client,
	}
}

func (c *Client) Pods() *pods.Client {
	return pods.NewClient(c.client)
}

func (c *Client) Images() *images.Client {
	return images.NewClient(c.client)
}
