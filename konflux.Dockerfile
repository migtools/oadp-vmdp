# Konflux hermetic build for the oadp-vmdp download server
# Dependencies are prefetched by the Konflux pipeline (Hermeto) and injected
# into the build context before this Containerfile runs.

FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_golang_1.25 AS builder

COPY . /workspace
WORKDIR /workspace

ENV GOEXPERIMENT=strictfipsruntime

# Version information
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown
ARG BUILDTAGS=

# Build oadp-vmdp binaries for all target platforms as direct executables
RUN mkdir -p /archives && \
    for platform in linux/amd64 linux/arm64 windows/amd64 windows/arm64; do \
        os=$(echo $platform | cut -d'/' -f1); \
        arch=$(echo $platform | cut -d'/' -f2); \
        if [ "$os" = "windows" ]; then \
            out_name="oadp-vmdp_${VERSION}_${os}_${arch}.exe"; \
        else \
            out_name="oadp-vmdp_${VERSION}_${os}_${arch}"; \
        fi; \
        echo "Building oadp-vmdp for ${os}/${arch}..."; \
        CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
            go build -trimpath -mod=mod \
            -tags="${BUILDTAGS}" \
            -ldflags="-s -w \
                -X github.com/kopia/kopia/repo.BuildVersion=${VERSION} \
                -X github.com/kopia/kopia/repo.BuildInfo=${BUILD_DATE}-${GIT_COMMIT} \
                -X github.com/kopia/kopia/repo.BuildGitHubRepo=github.com/openshift/oadp-vmdp" \
            -o /archives/$out_name \
            . ; \
    done && \
    cd /archives && sha256sum oadp-vmdp_* > sha256sum.txt && \
    rm -rf /root/.cache/go-build /tmp/*

# Build the download server (FIPS-compliant)
RUN CGO_ENABLED=1 GOOS=linux go build -mod=mod -a -tags strictfipsruntime \
    -o /workspace/bin/download-server ./cmd/downloads/ && \
    go clean -cache -modcache -testcache && \
    rm -rf /root/.cache/go-build /go/pkg

FROM registry.redhat.io/ubi9/ubi:latest

RUN dnf -y install openssl && dnf -y reinstall tzdata && dnf clean all

COPY --from=builder /archives /archives
COPY --from=builder /workspace/bin/download-server /usr/local/bin/download-server
COPY LICENSE /licenses/

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/download-server"]

LABEL description="OADP VMDP - Binary Download Server"
LABEL io.k8s.description="OADP VMDP - Binary Download Server"
LABEL io.k8s.display-name="OADP VMDP Downloads"
LABEL io.openshift.tags="oadp,backup,restore,virtualization,vmdp"
LABEL summary="Serves pre-built oadp-vmdp binaries for Linux and Windows"
