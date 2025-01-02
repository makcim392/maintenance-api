# Development stage
FROM golang:1.22-alpine AS dev

WORKDIR /app

RUN apk update && apk add --no-cache git

# Create a separate layer for air installation
RUN go install github.com/cosmtrek/air@v1.42.0

# Clean go cache and add the cleanup to the image
RUN go clean -cache -modcache

# Copy only dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Set PATH for air
ENV PATH="/go/bin:${PATH}"

# Set build flags to disable caching
ENV GOFLAGS="-buildvcs=false -a -gcflags='all=-N -l'"

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]