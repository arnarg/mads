# Pod named nginx
pod "nginx" {
  # Sets network config for the pod
  networking {
    # Map port 80 to 8080 on the host
    expose {
      container = 80
      host      = 8080
    }
  }

  # Container in nginx pod named nginx
  container "nginx" {
    # Specify image
    image = "docker.io/library/nginx:1.22.1"

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

  # Consul service named nginx
  # service "nginx" {
  #   # The port inside the container that the service
  #   # should map to
  #   port = 80

  #   # Service is connect enabled
  #   connect {
  #     # Inject a sidecar proxy
  #     sidecar_service {}
  #   }
  # }
}
