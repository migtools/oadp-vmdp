# OADP VM Data Protection (oadp-vmdp)

Virtual Machine Data Protection for OpenShift Virtualization.

OADP-VMDP is a command-line tool that runs inside virtual machines to back up and restore user data. It supports S3-compatible and filesystem storage backends.

---

## Supported Platforms

OADP-VMDP is built for [OpenShift Virtualization certified guest operating systems](https://access.redhat.com/articles/4234591) on x86_64 (amd64) and arm64 architectures.

---

## Quick Start

### 1. Create a Backup Storage Location (BSL)

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
