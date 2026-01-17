---
page_title: "trueform_pool Resource - Trueform"
subcategory: "Storage"
description: |-
  Manages a ZFS storage pool on TrueNAS.
---

# trueform_pool (Resource)

Manages a ZFS storage pool on TrueNAS Scale. Pools are the top-level storage containers in ZFS.

~> **Warning:** Destroying a pool will permanently delete all data stored in it. This action cannot be undone.

## Example Usage

### Basic Pool (Stripe)

```hcl
resource "trueform_pool" "storage" {
  name = "storage"

  topology = [
    {
      type  = "data"
      disks = ["sda", "sdb"]
    }
  ]
}
```

### Mirror Pool

```hcl
resource "trueform_pool" "tank" {
  name = "tank"

  topology = [
    {
      type  = "data"
      disks = ["sda", "sdb", "sdc", "sdd"]
    }
  ]
}
```

## Schema

### Required

- `name` (String) Name of the pool.
- `topology` (List of Object) Pool topology configuration.
  - `type` (String) Vdev type: `data`, `log`, `cache`, `spare`, `special`, `dedup`.
  - `disks` (List of String) List of disk identifiers.

### Optional

- `allow_duplicate_serials` (Boolean) Allow disks with duplicate serial numbers. Defaults to `false`.
- `checksum` (String) Checksum algorithm. Defaults to `on`.
- `deduplication` (String) Deduplication setting. Values: `ON`, `OFF`. Defaults to `OFF`.
- `encryption` (Boolean) Enable encryption. Defaults to `false`.

### Read-Only

- `allocated` (Number) Allocated space in bytes.
- `free` (Number) Free space in bytes.
- `healthy` (Boolean) Pool health status.
- `id` (Number) Pool identifier.
- `path` (String) Pool mount path.
- `size` (Number) Total pool size in bytes.
- `status` (String) Pool status (e.g., `ONLINE`, `DEGRADED`).

## Import

Pools can be imported using the pool name:

```shell
terraform import trueform_pool.tank tank
```
