# Basic nginx

This simple example is the equivalent of running of the following:

- `podman pod create -p 8080:80 nginx`
- `podman create --name nginx-nginx docker.io/library/nginx:1.22.1`
- `podman cp index.html nginx-nginx:/usr/share/nginx/html/index.html`
- `podman pod start nginx`

## Example

```yaml
name: nginx

containers:
  - name: nginx
    image: docker.io/library/nginx:1.22.1

    # Simple port mapping
    ports:
      - containerPort: 80
        hostPort: 8080

    # Mads can copy files into the pod before it starts
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
```

**Output:**

```
>> podman ps
CONTAINER ID  IMAGE                                   COMMAND               CREATED        STATUS            PORTS                 NAMES
d0adf09ec7c8  localhost/podman-pause:4.3.1-315532800                        7 seconds ago  Up 5 seconds ago  0.0.0.0:8080->80/tcp  3e04d72d17db-infra
0a7110683958  docker.io/library/nginx:1.22.1          nginx -g daemon o...  5 seconds ago  Up 5 seconds ago  0.0.0.0:8080->80/tcp  nginx-nginx

```
