# syntax=docker/dockerfile:1

FROM golang:1.24

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Create unprivileged system user and group (UID/GID 10001)
RUN groupadd --system --gid 10001 appgroup && \
    useradd --system --uid 10001 --gid appgroup --create-home --home-dir /home/appuser appuser

# Set up runtime directories with non-root ownership
RUN mkdir -p /data /app/data /home/appuser/.cache/go-build /home/appuser/go && \
    chown -R appuser:appgroup /data /app/data /home/appuser

# Copy source code and build binary
COPY . ./
RUN GOOS=linux go build -o /docker-go-solve && \
    chmod 755 /docker-go-solve

# Switch to unprivileged appuser for runtime execution
USER appuser
ENV HOME=/home/appuser \
    GOPATH=/home/appuser/go \
    GOCACHE=/home/appuser/.cache/go-build \
    DATABASE_PATH=/data/go-solve.db

# Pre-warm Go test cache as appuser so runtime code evaluations are fast
RUN GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0 go test -vet=off ./tasks/easy/hello_world ./tasks/easy/bubbles

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${PORT:-8080}/healthz || exit 1

CMD ["/docker-go-solve"]
