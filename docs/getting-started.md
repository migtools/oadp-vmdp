# OADP VMDP: A "Hello World" Getting Started Guide

## What is OADP VMDP?

**OADP VMDP** (VM Data Protection) is a CLI tool that runs **inside an OpenShift Virtualization guest VM** to perform **file-level backup and restore** to S3-compatible or filesystem storage. It complements OADP's existing snapshot-based full-VM backup — think of it as `rsync`-to-the-cloud for individual files and directories inside your VM, with encryption, deduplication, and incremental backups built in.

Under the hood, it is a rebranded [Kopia](https://kopia.io) client with a simplified command set.

---

## Prerequisites

Before you begin, you need:

- An **OpenShift cluster** with the **OADP Operator** installed (v1.6+)
- An **OpenShift Virtualization VM** running a supported guest OS (RHEL, Windows, etc.)
- **S3-compatible object storage** with a bucket and credentials (e.g., AWS S3, MinIO, Noobaa/ODF)
  - Bucket name, endpoint URL, access key, and secret key
- Network connectivity from the guest VM to your S3 endpoint

---

## Step 1: Download the `oadp-vmdp` Binary

The OADP Operator automatically deploys a **download server** inside the cluster.

### Option A: Download from the OpenShift Console

1. Log in to the OpenShift web console
2. Click the **"?"** (help) icon in the top navigation bar
3. Select **"Command line tools"**
4. Find **"oadp-vmdp - OADP VM Data Protection CLI"** in the list
5. Click **"Download OADP VMDP CLI"** — this downloads the binary for your platform

### Option B: Download via the Route URL directly

From inside your guest VM (or any machine with access to the route):

```bash
# Find the route URL (run this from a machine with oc access)
oc get route oadp-vmdp-server-route -n openshift-adp -o jsonpath='{.spec.host}'
```

Then, from inside your **guest VM**, fetch the binary:

```bash
# Replace <ROUTE_HOST> with the hostname from above
curl -kO https://<ROUTE_HOST>/download/oadp-vmdp_v1.0.0_linux_amd64
```

### Option C: Run the download server locally (for development/testing)

```bash
podman run --rm -p 8080:8080 quay.io/konveyor/oadp-vmdp-binaries:oadp-1.6
```

Then open `http://localhost:8080` and download the binary for your platform.

---

## Step 2: Install the Binary (inside your guest VM)

SSH into your OpenShift Virtualization VM and install:

**Linux:**

```bash
chmod +x oadp-vmdp_*_linux_amd64
sudo mv oadp-vmdp_*_linux_amd64 /usr/local/bin/oadp-vmdp

# Verify
oadp-vmdp --version
```

**Windows (PowerShell):**

```powershell
Rename-Item oadp-vmdp_*_windows_amd64.exe oadp-vmdp.exe
.\oadp-vmdp.exe --version
```

---

## Step 3: Create a Backup Storage Location (BSL)

This initializes an encrypted Kopia repository in your S3 bucket. You will be prompted for an **encryption password** — remember it, you'll need it to access your backups later.

```bash
oadp-vmdp bsl create s3 \
  --bucket my-backup-bucket \
  --endpoint s3.example.com \
  --access-key YOUR_ACCESS_KEY \
  --secret-access-key YOUR_SECRET_KEY
```

You'll be prompted:

```
Enter password to create new repository:
Re-enter password for verification:
```

Pick a strong password and save it somewhere safe. This encrypts all your backup data client-side before it ever leaves the VM.

> **Tip:** To skip the interactive prompt (useful for scripting), set `export BSLS_PASSWORD="your-secure-password"` before running the command.

Verify the connection:

```bash
oadp-vmdp bsl status
```

You should see output confirming you're connected to the repository.

> **Note:** OADP VMDP automatically prepends `oadp-vmdp/` as a prefix in your S3 bucket to keep its data isolated.

---

## Step 4: Create Some Test Data (the "Hello World")

Let's create a small directory with sample files to back up:

```bash
mkdir -p /tmp/hello-vmdp
echo "Hello from OADP VMDP!" > /tmp/hello-vmdp/greeting.txt
echo "This is my important config" > /tmp/hello-vmdp/app.conf
date > /tmp/hello-vmdp/timestamp.txt
```

---

## Step 5: Back Up Your Data

```bash
oadp-vmdp backup create /tmp/hello-vmdp
```

You should see output showing the files being scanned, deduplicated, encrypted, and uploaded. Something like:

```
Snapshotting user@my-vm:/tmp/hello-vmdp ...
 * 0 hashing, 3 hashed (150 B), 0 cached (0 B), uploaded 1.2 KB ...
Created snapshot with root ...
```

---

## Step 6: List Your Backups

```bash
oadp-vmdp backup list
```

This shows all snapshots you've created, with timestamps, source paths, and unique IDs.

---

## Step 7: Simulate Data Loss

Let's pretend disaster struck:

```bash
rm -rf /tmp/hello-vmdp
ls /tmp/hello-vmdp   # Should show: No such file or directory
```

---

## Step 8: Restore Your Data

Restore the most recent backup of `/tmp/hello-vmdp`:

```bash
oadp-vmdp restore /tmp/hello-vmdp
```

Or, to restore to a different location:

```bash
oadp-vmdp restore /tmp/hello-vmdp /tmp/restored-data
```

Verify the restore:

```bash
cat /tmp/hello-vmdp/greeting.txt
# Output: Hello from OADP VMDP!
```

Your data is back.

---

## Step 9: (Optional) Connect from Another VM

The same backups are accessible from a different VM. On the second VM, use `bsl connect` (not `bsl create`) with the **same bucket, credentials, and encryption password**:

```bash
oadp-vmdp bsl connect s3 \
  --bucket my-backup-bucket \
  --endpoint s3.example.com \
  --access-key YOUR_ACCESS_KEY \
  --secret-access-key YOUR_SECRET_KEY
```

Then list and restore:

```bash
oadp-vmdp backup list
oadp-vmdp restore /tmp/hello-vmdp /tmp/recovered-from-other-vm
```

---

## Quick Command Reference

| Task | Command |
|------|---------|
| Create BSL (new repo) | `oadp-vmdp bsl create s3 --bucket ... --endpoint ... --access-key ... --secret-access-key ...` |
| Connect to existing BSL | `oadp-vmdp bsl connect s3 --bucket ... --endpoint ... --access-key ... --secret-access-key ...` |
| Check BSL status | `oadp-vmdp bsl status` |
| Back up a path | `oadp-vmdp backup create /path/to/data` |
| List backups | `oadp-vmdp backup list` |
| Restore latest backup | `oadp-vmdp restore /path/to/data` |
| Restore specific backup | `oadp-vmdp restore <snapshot-id> /target/path` |
| Disconnect | `oadp-vmdp bsl disconnect` |
| Get help | `oadp-vmdp --help` |

---

## How It All Fits Together

```
+---------------------------------------------+
|  OpenShift Cluster                          |
|                                             |
|  OADP Operator                              |
|    |                                        |
|    +-- vmdp_download_controller.go          |
|    |     Creates:                           |
|    |       Deployment (binary server)       |
|    |       Service + Route                  |
|    |       ConsoleCLIDownload               |
|    |                                        |
|  +------+   Route    +------------------+  |
|  | User | ---------> | VMDP Download    |  |
|  |  VM  |  downloads | Server (nginx)   |  |
|  |      |  oadp-vmdp |                  |  |
|  +--+---+            +------------------+  |
|     |                                       |
+-----|---------------------------------------+
      |
      | oadp-vmdp bsl create / backup / restore
      |
      v
+------------------+
| S3 Object Storage|
| (encrypted data) |
+------------------+
```

---

## Key Concepts to Remember

- **`bsl create`** = initialize a *new* encrypted repository (first time only)
- **`bsl connect`** = reconnect to an *existing* repository (same bucket, same password)
- **`backup create`** = create a new snapshot (incremental by default)
- **`restore`** = pull data back from a snapshot
- All encryption happens **client-side inside the VM** before data leaves
- VMDP is **complementary** to full-VM snapshot backup via Velero/OADP — it handles file-level protection that users manage themselves

For the full reference, see the [README_OADP_VMDP.md](../README_OADP_VMDP.md) in the oadp-vmdp repository.
