---
page_title: "trueform_pool Data Source - Trueform"
subcategory: "Storage"
description: |-
  Retrieves information about an existing ZFS pool on TrueNAS.
---

# trueform_pool (Data Source)

Retrieves information about an existing ZFS storage pool on TrueNAS Scale.

## Example Usage

```hcl
data "trueform_pool" "tank" {
  name = "tank"
}

output "pool_status" {
  value = data.trueform_pool.tank.status
}

output "pool_free_space" {
  value = "${data.trueform_pool.tank.free / 1073741824} GB free"
}
```

## Schema

### Required

- `name` (String) Name of the pool to look up.

### Read-Only

- `allocated` (Number) Allocated space in bytes.
- `free` (Number) Free space in bytes.
- `healthy` (Boolean) Pool health status.
- `id` (Number) Pool identifier.
- `path` (String) Pool mount path.
- `size` (Number) Total pool size in bytes.
- `status` (String) Pool status (e.g., `ONLINE`, `DEGRADED`).
