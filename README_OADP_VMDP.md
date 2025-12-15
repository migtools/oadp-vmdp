# OADP VM Data Protection (oadp-vmdp)

Virtual Machine Data Protection for OpenShift Virtualization.

OADP-VMDP is a command-line tool that runs inside virtual machines to back up and restore user data. It supports S3-compatible and filesystem storage backends.

---

## Supported Platforms

OADP-VMDP is built for [OpenShift Virtualization certified guest operating systems](https://access.redhat.com/articles/4234591) on x86_64 (amd64) and arm64 architectures.

---

## Quick Start

### 1. Create a Backup Storage Location (BSL)

If using aws:
** note --endpoint w/ "s3.<region>.amazonaws.com" did not work

```bash
oadp-vmdp bsl create s3 \
  --bucket my-backup-bucket \
  --access-key YOUR_ACCESS_KEY \
  --secret-access-key YOUR_SECRET_KEY
```

```bash
oadp-vmdp bsl create s3 \
  --bucket my-backup-bucket \
  --endpoint s3.example.com \
  --access-key YOUR_ACCESS_KEY \
  --secret-access-key YOUR_SECRET_KEY
```

### 2. Create a Backup

```bash
oadp-vmdp backup create /path/to/data
```

### 3. Restore from Backup

```bash
oadp-vmdp restore /path/to/data
```

---

## Commands

### BSL (Backup Storage Location)

| Command | Description |
|---------|-------------|
| `bsl create` | Create and connect to a new BSL |
| `bsl connect` | Connect to an existing BSL |
| `bsl disconnect` | Disconnect from current BSL |
| `bsl status` | Show current BSL connection status |
| `bsl change-password` | Change the BSL encryption password |

### Backup

| Command | Description |
|---------|-------------|
| `backup create` | Create a new backup of specified path(s) |
| `backup list` | List all available backups |
| `backup delete` | Delete a specific backup |
| `restore` | Restore data from a backup |

---

## Storage Backends

For certified providers, see [OADP Certified Backup Storage Providers](https://docs.redhat.com/en/documentation/openshift_container_platform/latest/html/backup_and_restore/oadp-application-backup-and-restore#oadp-certified-backup-storage-providers_about-installing-oadp).

### S3-Compatible Storage

| Option | Description | Default |
|--------|-------------|---------|
| `--bucket` | Name of the S3 bucket | (required) |
| `--access-key` | Access Key ID | (required) |
| `--secret-access-key` | Secret Access Key | (required) |
| `--endpoint` | S3 endpoint URL | `s3.amazonaws.com` |
| `--region` | S3 region | (auto-detect) |
| `--prefix` | Object prefix in bucket | (none) |
| `--session-token` | Session token for temporary credentials | (none) |
| `--disable-tls` | Disable HTTPS | `false` |
| `--disable-tls-verification` | Skip TLS certificate verification | `false` |
| `--root-ca-pem-path` | Path to custom CA certificate file | (none) |
| `--root-ca-pem-base64` | Base64-encoded CA certificate | (none) |

> **Note:** OADP-VMDP automatically prepends `oadp-vmdp/` to your prefix.

### Filesystem Storage

| Option | Description | Default |
|--------|-------------|---------|
| `--path` | Absolute path to storage directory | (required) |
| `--owner-uid` | User ID for new files | (current user) |
| `--owner-gid` | Group ID for new files | (current group) |
| `--file-mode` | Permission mode for files | `0600` |
| `--dir-mode` | Permission mode for directories | `0700` |

---

## Environment Variables

### Credentials

| Variable | Description |
|----------|-------------|
| `BSLS_PASSWORD` | BSL encryption password (avoids interactive prompt) |
| `AWS_ACCESS_KEY_ID` | Access key for S3 storage |
| `AWS_SECRET_ACCESS_KEY` | Secret key for S3 storage |
| `AWS_SESSION_TOKEN` | Session token for temporary credentials |

### Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `OADP_CONFIG_PATH` | Path to configuration file | `~/.config/oadp/repository.config` |
| `OADP_CACHE_DIRECTORY` | Path to cache directory | (system dependent) |
| `OADP_LOG_DIR` | Directory for log files | `~/.cache/oadp/` |

### Behavior

| Variable | Description | Default |
|----------|-------------|---------|
| `OADP_CHECK_FOR_UPDATES` | Enable/disable update checks | `true` |
| `OADP_PERSIST_CREDENTIALS_ON_CONNECT` | Save credentials after connecting | `true` |
| `OADP_USE_KEYRING` | Use system keyring for password storage | `false` |
| `OADP_BACKUP_FAIL_FAST` | Fail immediately on first error | `false` |

### Logging

| Variable | Description | Default |
|----------|-------------|---------|
| `OADP_LOG_DIR_MAX_FILES` | Maximum number of log files | `1000` |
| `OADP_LOG_DIR_MAX_AGE` | Maximum age of log files | `720h` |
| `OADP_LOG_DIR_MAX_SIZE_MB` | Maximum total size of log files (MB) | `1000` |

---

## Workflows

### Non-Interactive Usage (Scripts/Automation)

Set credentials via environment variables to avoid interactive prompts:

```bash
export BSLS_PASSWORD="your-secure-password"
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"

oadp-vmdp bsl create s3 --bucket my-bucket --endpoint s3.example.com
oadp-vmdp backup create /path/to/data
```

### Connecting from Another System

To access backups from a different VM, use `bsl connect` instead of `bsl create`:

```bash
oadp-vmdp bsl connect s3 \
  --bucket my-backup-bucket \
  --endpoint s3.example.com \
  --access-key YOUR_ACCESS_KEY \
  --secret-access-key YOUR_SECRET_KEY
```

### Restoring a Specific Backup

```bash
# List available backups
oadp-vmdp backup list

# Restore a specific backup by ID to a custom location
oadp-vmdp restore <source-id> /path/to/restore/
```

---

## File Locations

| Type | Linux | Windows |
|------|-------|---------|
| Configuration | `~/.config/oadp/repository.config` | `%APPDATA%\oadp\repository.config` |
| Logs | `~/.cache/oadp/` | `%LOCALAPPDATA%\oadp\` |

---

## Troubleshooting

### "Not connected to a Backup Storage Location"

```bash
oadp-vmdp bsl status    # Check current status
oadp-vmdp bsl connect   # Connect to existing BSL
```

### "prefix must not contain 'oadp-vmdp'"

The `oadp-vmdp/` prefix is added automatically. Don't include `oadp-vmdp` as a path segment in `--prefix`.
Also ensure your `--prefix` does not start or end with whitespace.

### S3 Connection Issues

For self-hosted S3-compatible services, you may need:

- `--disable-tls` for non-HTTPS endpoints
- `--disable-tls-verification` for self-signed certificates
- `--root-ca-pem-path` to specify a custom CA certificate

---

## Getting Help

```bash
oadp-vmdp --help
oadp-vmdp bsl --help
oadp-vmdp backup create --help
```

---

## Download Server (ConsoleCLIDownload)

OADP-VMDP includes a download server that runs inside an OpenShift cluster and serves pre-built binaries for KubeVirt guest VMs. Users can download the correct binary for their VM's guest operating system directly from the OpenShift console or via HTTP.

Supported guest operating systems:
- **Red Hat Enterprise Linux** (x86_64, aarch64)
- **Microsoft Windows** (x86_64, aarch64)

Each binary is statically linked and includes a SHA256 checksum for integrity verification.

This is powered by:
- **`cmd/downloads/server.go`** - A lightweight Go HTTP server that serves the binaries from `/archives`
- **`Containerfile.download`** - Builds all platform binaries with SHA256 checksums and packages them with the download server
- **`.github/workflows/quay_binaries_push.yml`** - CI workflow that builds and pushes the image to `quay.io/konveyor/oadp-vmdp-binaries`

### Building the Download Server Locally

```bash
# Build with Podman (builds for amd64 and arm64 container platforms)
make -f Makefile.ubi container-download-build

# Build and push
make -f Makefile.ubi container-download-build-push \
  DOWNLOAD_IMAGE=quay.io/youruser/oadp-vmdp-binaries TAG=dev
```

### Running the Download Server Locally

The easiest way is to pull the pre-built image from Quay:

```console
$ podman run --rm -p 8080:8080 quay.io/konveyor/oadp-vmdp-binaries:latest
```

Then open http://localhost:8080 in your browser to see the download page with all available binaries and their SHA256 checksums.

Alternatively, build the image from source:

```console
$ podman build \
    --build-arg TARGETOS=linux \
    --build-arg TARGETARCH=amd64 \
    --build-arg VERSION=dev \
    -t oadp-vmdp-binaries:dev \
    -f Containerfile.download .

$ podman run --rm -p 8080:8080 oadp-vmdp-binaries:dev
```

### Downloading and Verifying Binaries

**Red Hat Enterprise Linux (x86_64):**

```console
$ curl -O http://localhost:8080/download/oadp-vmdp_v1.0.0_linux_amd64
$ curl -O http://localhost:8080/download/sha256sum.txt
$ sha256sum -c sha256sum.txt
oadp-vmdp_v1.0.0_linux_amd64: OK
$ chmod +x oadp-vmdp_v1.0.0_linux_amd64
$ sudo mv oadp-vmdp_v1.0.0_linux_amd64 /usr/local/bin/oadp-vmdp
$ oadp-vmdp --version
```

**Red Hat Enterprise Linux (aarch64):**

```console
$ curl -O http://localhost:8080/download/oadp-vmdp_v1.0.0_linux_arm64
$ curl -O http://localhost:8080/download/sha256sum.txt
$ sha256sum -c sha256sum.txt
oadp-vmdp_v1.0.0_linux_arm64: OK
$ chmod +x oadp-vmdp_v1.0.0_linux_arm64
$ sudo mv oadp-vmdp_v1.0.0_linux_arm64 /usr/local/bin/oadp-vmdp
$ oadp-vmdp --version
```

**Microsoft Windows (x86_64) - PowerShell:**

```powershell
PS> Invoke-WebRequest -Uri http://localhost:8080/download/oadp-vmdp_v1.0.0_windows_amd64.exe -OutFile oadp-vmdp.exe
PS> Invoke-WebRequest -Uri http://localhost:8080/download/sha256sum.txt -OutFile sha256sum.txt
PS> (Get-FileHash oadp-vmdp.exe -Algorithm SHA256).Hash
PS> Select-String -Path sha256sum.txt -Pattern "windows_amd64"
PS> .\oadp-vmdp.exe --version
```

**Microsoft Windows (aarch64) - PowerShell:**

```powershell
PS> Invoke-WebRequest -Uri http://localhost:8080/download/oadp-vmdp_v1.0.0_windows_arm64.exe -OutFile oadp-vmdp.exe
PS> Invoke-WebRequest -Uri http://localhost:8080/download/sha256sum.txt -OutFile sha256sum.txt
PS> (Get-FileHash oadp-vmdp.exe -Algorithm SHA256).Hash
PS> Select-String -Path sha256sum.txt -Pattern "windows_arm64"
PS> .\oadp-vmdp.exe --version
```

### How It Works in OpenShift

1. The OADP operator deploys the download server container (via `RELATED_IMAGE_VMDP_CLI_DOWNLOAD` env var)
2. A `ConsoleCLIDownload` resource is created, linking to the download server's routes
3. Users see download links in the OpenShift console and can fetch the binary matching their guest OS
4. Each binary is statically linked and includes a SHA256 checksum - download, verify, and run

---

## Kopia Compatibility

OADP-VMDP is based on [Kopia](https://kopia.io) and uses the same repository format. Repositories are fully compatible between the two tools.

**Command mapping:**

| oadp-vmdp | kopia |
|-----------|-------|
| `bsl` | `repository` |
| `backup` | `snapshot` |

**Using Kopia CLI with oadp-vmdp repositories:**

When connecting with Kopia CLI, include the `oadp-vmdp/` prefix that oadp-vmdp adds automatically:

```bash
kopia repository connect s3 \
  --bucket my-bucket \
  --prefix oadp-vmdp/my-prefix/ \
  ...
```

**Using oadp-vmdp with existing Kopia repositories:**

oadp-vmdp will prepend `oadp-vmdp/` to your prefix. To access an existing Kopia repository at prefix `backups/`, you cannot connect directly - the prefix manipulation would cause a mismatch.

---

## License

OADP-VMDP is based on Kopia and is distributed by Red Hat, Inc.
