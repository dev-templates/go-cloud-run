# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:latest as builder

# Create and change to the app directory.
WORKDIR /go/src/github.com/app/server

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -ldflags="-w -s" -v -o /app/server ./cmd/go-cloud-run

# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates && rm -rf /var/cache/apk/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server

# Run the web service on container startup.
CMD ["/app/server"]