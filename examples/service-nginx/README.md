# Service nginx

This example integrates with consul to create a service for the pod, with consul connect sidecar definition and automatically injects an envoy sidecar proxy into the pod. This is equivalent to:

- `consul services register nginx.json` (where nginx.json is an equivalent service definition).
- `podman pod create -p {{sidecar_service_port}}:{{sidecar_service_port}} nginx` (where `sidecar_service_port` is a port dynamically provisioned by consul).
- `podman create --name nginx-nginx docker.io/library/nginx:1.22.1`
- `podman cp index.html nginx-nginx:/usr/share/nginx/html/index.html`
- `podman create --name nginx-nginx-sidecar-proxy envoyproxy/envoy -c /etc/envoy/envoy.json`
- `consul connect envoy -sidecar-for nginx -bootstrap > envoy-config.json`
- `podman cp envoy-config.json nginx-nginx-sidecar-proxy:/etc/envoy/envoy.json`
- `podman pod start nginx`

## Example

```yaml
name: nginx

containers:
  - name: nginx
    image: docker.io/library/nginx:1.22.1
    files:
      - destination: /usr/share/nginx/html/index.html
        content: |
          <!DOCTYPE html>
          <html>
          <head>
          <title>Welcome to mads!</title>
          <style>
          html { color-scheme: light dark; }
          body { width: 35em; margin: 0 auto;
          font-family: Tahoma, Verdana, Arial, sans-serif; }
          </style>
          </head>
          <body>
            <h1>Welcome to mads!</h1>
          </body>
          </html>

services:
  - name: nginx
    port: 80
    connect:
      sidecarService: {}
```
