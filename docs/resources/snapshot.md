---
page_title: "trueform_snapshot Resource - Trueform"
subcategory: "Storage"
description: |-
  Manages a ZFS snapshot on TrueNAS.
---

# trueform_snapshot (Resource)

Manages a ZFS snapshot on TrueNAS Scale. Snapshots provide point-in-time copies of datasets for backup and recovery.

~> **Note:** Snapshots are immutable. Changing the `name` or `dataset` will force recreation of the snapshot.

## Example Usage

### Basic Snapshot

```hcl
resource "trueform_snapshot" "daily" {
  dataset = "tank/data"
  name    = "daily-backup"
}
```

### Recursive Snapshot

```hcl
resource "trueform_snapshot" "full_backup" {
  dataset   = "tank"
  name      = "full-backup"
  recursive = true
}
```

## Schema

### Required

- `dataset` (String) Dataset to snapshot (pool/dataset format).
- `name` (String) Snapshot name.

### Optional

- `recursive` (Boolean) Create snapshots recursively for child datasets. Defaults to `false`.
- `vmware_sync` (Boolean) VMware sync for consistent VM snapshots. Defaults to `false`.

### Read-Only

- `creation_time` (String) Snapshot creation timestamp.
- `holds` (List of String) List of holds on the snapshot.
- `id` (String) Snapshot identifier (dataset@name format).
- `referenced_bytes` (Number) Referenced data size in bytes.
- `used_bytes` (Number) Space used by snapshot in bytes.

## Import

Snapshots can be imported using the dataset@name format:

```shell
terraform import trueform_snapshot.daily "tank/data@daily-backup"
```
