---
page_title: "trueform_iscsi_extent Resource - Trueform"
subcategory: "iSCSI"
description: |-
  Manages an iSCSI extent (LUN) on TrueNAS.
---

# trueform_iscsi_extent (Resource)

Manages an iSCSI extent on TrueNAS Scale. Extents represent the actual storage (disk or file) that is exposed as a LUN.

## Example Usage

### File-based Extent

```hcl
resource "trueform_iscsi_extent" "data_lun" {
  name     = "data-lun0"
  type     = "FILE"
  path     = "/mnt/tank/iscsi/data.img"
  filesize = 107374182400  # 100GB

  blocksize = 512
  rpm       = "SSD"
  comment   = "Data LUN"
}
```

### Zvol-based Extent

```hcl
resource "trueform_iscsi_extent" "vm_disk" {
  name = "vm-disk"
  type = "DISK"
  disk = "zvol/tank/iscsi/vm-disk"

  blocksize = 4096
  rpm       = "SSD"
}
```

## Schema

### Required

- `name` (String) Extent name. Cannot be changed after creation.
- `type` (String) Extent type. Values: `FILE`, `DISK`. Cannot be changed after creation.

### Optional (for FILE type)

- `path` (String) Path for file-based extent.
- `filesize` (Number) Size of file extent in bytes.

### Optional (for DISK type)

- `disk` (String) Zvol path for disk-based extent.

### Optional

- `avail_threshold` (Number) Alert threshold for available space percentage.
- `blocksize` (Number) Logical block size. Values: `512`, `4096`. Defaults to `512`.
- `comment` (String) Extent description.
- `enabled` (Boolean) Enable the extent. Defaults to `true`.
- `insecure_tpc` (Boolean) Allow Third Party Copy operations. Defaults to `true`.
- `pblocksize` (Boolean) Use physical block size. Defaults to `false`.
- `ro` (Boolean) Read-only extent. Defaults to `false`.
- `rpm` (String) Reported RPM. Values: `UNKNOWN`, `SSD`, `5400`, `7200`, `10000`, `15000`. Defaults to `SSD`.
- `xen` (Boolean) Xen compatibility mode. Defaults to `false`.

### Read-Only

- `id` (Number) Extent identifier.
- `locked` (Boolean) Whether the extent is locked.
- `naa` (String) NAA identifier.
- `serial` (String) Serial number.

## Import

iSCSI extents can be imported using the extent ID:

```shell
terraform import trueform_iscsi_extent.data_lun 1
```
