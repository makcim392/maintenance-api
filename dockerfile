FROM golang:1.21-alpine

WORKDIR /app

RUN apk update && apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

# Use a pinned version that still has "module github.com/cosmtrek/air"
RUN go install github.com/cosmtrek/air@v1.42.0

ENV PATH="/go/bin:${PATH}"

COPY . .

EXPOSE 8080
CMD ["air"]
