# Copyright 2025 Red Hat Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ==============================================================================
# OADP VM Data Protection - Multi-Architecture Container Build
# ==============================================================================
#
# This Containerfile builds a statically-linked oadp-vmdp CLI binary for Linux
# and packages it in a minimal UBI container image.
#
# Supported architectures:
#   - linux/amd64 (x86_64)
#   - linux/arm64 (aarch64)
#
# Usage with Docker:
#   docker buildx build --platform linux/amd64,linux/arm64 -t oadp-vmdp .
#
# Usage with Podman:
#   podman build --arch amd64 -t oadp-vmdp:amd64 .
#   podman build --arch arm64 -t oadp-vmdp:arm64 .
#   podman build --arch amd64 \
#     --build-arg VERSION=1.0.0 \
#     --build-arg GIT_COMMIT=5eaa13d1 \
#     --build-arg BUILD_DATE=2025-12-15T00:00:00Z \
#     --build-arg BUILDTAGS=oadp \
#     -t oadp-vmdp:amd64 .
#

# ==============================================================================
# Build Stage - Compile the Go binary
# ==============================================================================

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

# Build arguments for cross-compilation
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH

# Version information (passed from Makefile.ubi)
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown
ARG BUILDTAGS=

# Install git for version detection (if not passed via args)
RUN apk add --no-cache git

WORKDIR /build

# Copy Go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with static linking
# CGO_ENABLED=0 ensures a fully static binary that works on any Linux distro
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
        -trimpath \
        -mod=mod \
        -tags="${BUILDTAGS}" \
        -ldflags="-s -w \
            -X github.com/kopia/kopia/repo.BuildVersion=${VERSION} \
            -X github.com/kopia/kopia/repo.BuildInfo=${BUILD_DATE}-${GIT_COMMIT} \
            -X github.com/kopia/kopia/repo.BuildGitHubRepo=github.com/openshift/oadp-vmdp" \
        -o /build/oadp-vmdp \
        .

# ==============================================================================
# Runtime Stage - Minimal container with just the binary
# ==============================================================================

FROM registry.access.redhat.com/ubi9-minimal:latest

# Version information (re-declared for this stage so LABEL can use them)
# NOTE: ARG scope does not automatically carry across FROM boundaries.
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown
ARG BUILDTAGS=

# Labels for container metadata
LABEL name="oadp-vmdp" \
      vendor="Red Hat, Inc." \
      version="${VERSION}" \
      summary="OADP VM Data Protection CLI" \
      description="Virtual Machine Data Protection tool for OpenShift Virtualization backup and restore operations" \
      io.k8s.display-name="OADP VM Data Protection" \
      io.k8s.description="CLI tool for VM data protection in OpenShift Virtualization" \
      io.openshift.tags="oadp,backup,restore,virtualization"

# Create cache directory for the application
# This is needed by kopia/oadp-vmdp for caching operations
WORKDIR /

RUN mkdir -p /.cache && \
    chown 65532:65532 /.cache

# Copy the binary from builder stage
COPY --from=builder /build/oadp-vmdp /oadp-vmdp

# Run as non-root user for security
USER 65532:65532

# Set the entrypoint
ENTRYPOINT ["/oadp-vmdp"]
