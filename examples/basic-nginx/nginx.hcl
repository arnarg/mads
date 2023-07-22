# Container named nginx
container "nginx" {
  # Specify image
  image = "docker.io/library/nginx:1.22.1"

  # Set environment variables
  env {
    NGINX_HOST = "example.com"
    NGINX_PORT = 80
  }

  # Sets metadata for the container
  metadata {
    # Adds annotation "io.mads/some-annotation" to container
    annotation "io.mads/some-annotation" {
      value = true
    }

    # Adds label "mads" to the container
    label "created_by" {
      value = "mads"
    }
  }

  # Sets network config for the container
  networking {
    # Adds a host mapping to container
    host "some.host.com" {
      address = "127.0.0.1"
    }

    # Map port 80 to 8080 on the host
    expose {
      container = 80
      host      = 8080
    }
  }

  # Bind mount into container
  mount {
    type        = "bind"
    source      = "/tmp/data"
    destination = "/var/lib/data"
  }

  # Write a file into the container before
  # starting it
  file {
    destination = "/usr/share/nginx/html/index.html"
    content     = <<-EOH
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
    EOH
  }
}
