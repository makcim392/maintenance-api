# Description

This application manages maintenance tasks performed during a working day. It provides a REST API for two types of users:
- **Technicians**: Can create, view, and update their own tasks
- **Managers**: Can view all tasks and delete them

Each task contains:
- A summary (max 2500 characters)
- A date when it was performed (performed_at)
- The technician who performed it

## API Endpoints

### Authentication
- **POST /register**
    - Registers a new user
    - Request body:
      ```json
      {
        "username": "string",
        "password": "string",
        "role": "technician|manager"
      }
      ```

- **POST /login**
    - Authenticates a user and returns a JWT token
    - Request body:
      ```json
      {
        "username": "string",
        "password": "string"
      }
      ```

### Tasks
- **POST /tasks**
    - Creates a new task
    - Requires authentication (Bearer token)
    - Request body:
      ```json
      {
        "summary": "Task description (max 2500 chars)",
        "performed_at": "2024-12-29T10:30:00Z"
      }
      ```

- **GET /tasks**
    - Lists tasks
    - Requires authentication (Bearer token)
    - Technicians: Returns only their tasks
    - Managers: Returns all tasks

- **PUT /tasks/{task_id}**
    - Updates an existing task
    - Requires authentication (Bearer token)
    - Only available to the technician who created the task
    - Request body:
      ```json
      {
        "summary": "Updated task description",
        "performed_at": "2024-12-29T10:30:00Z"
      }
      ```

- **DELETE /tasks/{task_id}**
    - Deletes a task
    - Requires authentication (Bearer token)
    - Only available to managers

# Running the project

1. Run `docker-compose up -d` to start the containers

# Building/rebuilding the image

In order to build the image, and considering docker build is getting deprecated, we can use:

```bash
docker buildx build --load -t sword_app:v1 . 
```

# Getting Started

### Pre-loaded Data
The application comes with pre-loaded data for testing purposes:

Users:
- `manager1` (Manager role)
- `john_tech` (Technician role)
- `sarah_manager` (Manager role)
- `makcim` (Technician role)

Several maintenance tasks are also pre-loaded into the database to demonstrate the application's functionality.

### API Documentation
A Postman collection is included in the `docs/postman` directory. To use it:

1. Import the collection into Postman
2. The collection includes requests for all available endpoints
3. Environment variables are set up for both local and development environments
4. Test credentials are included in the collection for both manager and technician roles

> Note: The pre-loaded data and Postman collection are intended for development and testing purposes only. 
> In a production environment, you should remove the test data and use secure credentials.

# Database access

A default.env file is provided in order to set the environment variables for the database connection. 
You can copy it to .env and modify the values as needed.

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

## Debugging

In order to debug the application, change the `APP_ENV` variable in the .env file to `dev` and run 
docker with `docker-compose up mysql`

## Make

A Makefile is provided to simplify common tasks, included linting and running tests as well as showing test coverage.

# Tests

The app has 2 types of tests:

- Unit tests: These tests are located in the `internal` directory and test individual functions and methods.
- Integration tests: These tests are located in the `tests` directory and test the application as a whole, 
- including the database and HTTP endpoints.

### Removing integration containers

```bash
docker-compose -f tests/docker-compose.test.yml down -v
```

# Future Improvements

## API Documentation 
Implement OpenAPI/Swagger documentation to provide interactive API exploration and make the endpoints more discoverable. 
This would include request/response schemas, examples, and authentication requirements, 
enabling faster integration for other developers.

## Microservices partitioning
The application could be further divided into microservices, for example, separating the user management functionality
(login and registration) from the rest of the application.


