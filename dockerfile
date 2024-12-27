# Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

# Pre-copy/cache go.mod for pre-downloading dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

EXPOSE 8080
CMD ["./main"]