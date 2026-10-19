# syntax=docker/dockerfile:1

FROM golang:1.24

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY . ./

# Build
RUN GOOS=linux go build -o /docker-go-solve

# Pre-warm Go test runner cache so runtime executions are fast
RUN GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0 go test -vet=off ./tasks/easy/hello_world ./tasks/easy/bubbles

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 8080

# Run
CMD ["/docker-go-solve"]
