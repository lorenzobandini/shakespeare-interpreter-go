# Docker Infrastructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development or executing-plans.
>
> **Goal:** Upgrade the Dockerfile to a 3-stage build (toolchain → check → runtime), update .dockerignore to allow tooling configs into the build context, and add docker:build / docker:check tasks to the Taskfile.
>
> **Architecture:** Multi-stage Dockerfile where stage 1 (	oolchain) installs all pinned dev tools and downloads Go deps, stage 2 (check) runs the full 	ask check suite, and stage 3 (untime) produces a minimal Alpine image with only the shpl binary.
>
> **Tech Stack:** Docker (multi-stage), Go 1.26.5-alpine base, Alpine 3.19 runtime, Task v3.42.1, golangci-lint v2.12.2.

## Global Constraints

- Go 1.26.5 (image tag golang:1.26.5-alpine).
- Runtime base lpine:3.19 (match existing Dockerfile).
- Tool versions pinned to exact semver (no @latest).
- 	ask check is the single CI gate — the check stage must run exactly this.
- All changes are backward-compatible: existing build workflows (	ask build, CI) unaffected.
- Working dir: C:\Users\lboa\Desktop\shakespeare-interpreter-go

## Task 1: Update .dockerignore

**Files:** Modify: .dockerignore

Replace current content (which ignores Taskfile.yaml and .golangci.yaml) with content that passes those files through to the build context.

**Current content:**
`
bin/
.git/
coverage.out
Taskfile.yaml
.golangci.yaml
`

**New content:**
`
bin/
.git/
.github/
docs/
coverage.out
`

## Task 2: Rewrite Dockerfile to 3-stage build

**Files:** Modify: Dockerfile (full rewrite)

`dockerfile
FROM golang:1.26.5-alpine AS toolchain

RUN apk add --no-cache curl ca-certificates git

WORKDIR /app

# Install Task
RUN go install github.com/go-task/task/v3/cmd/task@v3.42.1

# Install golangci-lint v2.12.2
RUN curl -sSfL https://github.com/golangci/golangci-lint/releases/download/v2.12.2/golangci-lint-2.12.2-linux-amd64.tar.gz \
    | tar -xz -C /usr/local/bin --strip=1 golangci-lint-2.12.2-linux-amd64/golangci-lint

# Install govulncheck + goimports
RUN go install golang.org/x/vuln/cmd/govulncheck@v1.2.3 \
    && go install golang.org/x/tools/cmd/goimports@v0.30.0

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy everything else (source, Taskfile.yaml, .golangci.yaml)
COPY . .

# Build the static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o shpl cmd/shpl/main.go

# Stage 2: check
FROM toolchain AS check
RUN task check

# Stage 3: runtime
FROM alpine:3.19 AS runtime
WORKDIR /root/
COPY --from=toolchain /app/shpl .
ENTRYPOINT ["./shpl"]
CMD ["--help"]
`

## Task 3: Add docker:build and docker:check tasks to Taskfile.yaml

**Files:** Modify: Taskfile.yaml

Add two new tasks after the existing uild block:

`yaml
  docker:build:
    desc: Build the minimal runtime Docker image
    cmds:
      - docker build -t shpl:latest .

  docker:check:
    desc: Run the full quality gate suite inside the container
    cmds:
      - docker build --target=check -t shpl:check .
`

## Task 4: Runtime image smoke tests

**Files:** None modified.

Verify shpl:latest image works:
- docker run --rm shpl:latest --help → CLI help
- docker run --rm shpl:latest version → version string
- docker run --rm -v "%cd%/testdata:/testdata" shpl:latest run /testdata/interpreter/hello.shpl → STX byte

