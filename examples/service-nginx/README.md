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

**Output:**

```
>> podman ps
CONTAINER ID  IMAGE                                   COMMAND               CREATED        STATUS            PORTS                     NAMES
9edff0205ebc  localhost/podman-pause:4.3.1-315532800                        7 seconds ago  Up 4 seconds ago  0.0.0.0:21000->21000/tcp  1509804e0a7c-infra
4d05d1d3648c  docker.io/library/nginx:1.22.1          nginx -g daemon o...  5 seconds ago  Up 4 seconds ago  0.0.0.0:21000->21000/tcp  nginx-nginx
f24457b82105  docker.io/envoyproxy/envoy:v1.22.8      -c /etc/envoy/env...  4 seconds ago  Up 4 seconds ago  0.0.0.0:21000->21000/tcp  nginx-nginx-sidecar-proxy

>> consul catalog services
consul
nginx
nginx-sidecar-proxy
```
