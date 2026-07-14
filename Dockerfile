# Stage 1: toolchain — pinned dev tools + source + dependencies
FROM golang:1.26.5-alpine AS toolchain

RUN apk add --no-cache curl ca-certificates git build-base

WORKDIR /app

# Install Task
RUN go install github.com/go-task/task/v3/cmd/task@v3.42.1

# Install golangci-lint v2.12.2
RUN curl -sSfL https://github.com/golangci/golangci-lint/releases/download/v2.12.2/golangci-lint-2.12.2-linux-amd64.tar.gz \
    | tar -xz -C /usr/local/bin --strip=1 golangci-lint-2.12.2-linux-amd64/golangci-lint

# Install govulncheck + goimports
RUN go install golang.org/x/vuln/cmd/govulncheck@v1.5.0 \
    && go install golang.org/x/tools/cmd/goimports@v0.30.0

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy everything else (source, Taskfile.yaml, .golangci.yaml)
COPY . .

# Build the static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o spl cmd/spl/main.go

# Stage 2: check — run the full CI gate
FROM toolchain AS check
RUN task check

# Stage 3: runtime — minimal image
FROM alpine:3.24 AS runtime
WORKDIR /root/
COPY --from=toolchain /app/spl .
ENTRYPOINT ["./spl"]
CMD ["--help"]
