# Description

# Running the project

1. Run `docker-compose up -d` to start the containers

# Building/rebuilding the image

In order to build the image, and considering docker build is getting deprecated, we can use:

`docker buildx build --load -t sword_app:v1 . `

# Database access

A default.env file is provided in order to set the environment variables for the database connection. You can copy it to .env and modify the values as needed.

# Developing

## Local Development (Hot-Reloading)

When actively coding, you’ll likely want immediate feedback on code changes without rebuilding the Docker image each time. To achieve this:

    Use volume mounts: Map your local project folder into the container so that any changes on your machine instantly reflect inside the container.
    Use a hot-reload tool (e.g., Air or CompileDaemon): The tool automatically watches for file changes and rebuilds or restarts the Go application.

A typical docker-compose.yml snippet for development might look like:

```docker
version: '3.8'

services:
  app:
    build: .
    volumes:
      - ./:/app        # Maps local code to /app in the container
    ports:
      - "8080:8080"
    command: ["air"]   # Hot-reload command (if using Air)

    volumes: - ./:/app means “take everything from the current directory (where docker-compose.yml lives) and mount it to /app in the container.”
    command: ["air"] (or similar) runs your hot-reload tool, which watches for changes and re-compiles the Go code automatically.

```

## Make

A Makefile is provided to simplify common tasks, included linting and running tests as well as showing test coverage.