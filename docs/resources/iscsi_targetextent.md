---
page_title: "trueform_iscsi_targetextent Resource - Trueform"
subcategory: "iSCSI"
description: |-
  Manages an iSCSI target-extent mapping on TrueNAS.
---

# trueform_iscsi_targetextent (Resource)

Manages an iSCSI target-extent mapping on TrueNAS Scale. This resource connects an extent (LUN) to a target, making the storage accessible to initiators.

## Example Usage

### Basic Mapping

```hcl
resource "trueform_iscsi_targetextent" "lun0" {
  target = trueform_iscsi_target.storage.id
  extent = trueform_iscsi_extent.data_lun.id
  lunid  = 0
}
```

### Multiple LUNs on Same Target

```hcl
resource "trueform_iscsi_targetextent" "data_lun" {
  target = trueform_iscsi_target.storage.id
  extent = trueform_iscsi_extent.data.id
  lunid  = 0
}

resource "trueform_iscsi_targetextent" "backup_lun" {
  target = trueform_iscsi_target.storage.id
  extent = trueform_iscsi_extent.backup.id
  lunid  = 1
}
```

## Schema

### Required

- `extent` (Number) Extent ID to map.
- `lunid` (Number) LUN ID (0-1023).
- `target` (Number) Target ID to map to.

### Read-Only

- `id` (Number) Mapping identifier.

## Import

iSCSI target-extent mappings can be imported using the mapping ID:

```shell
terraform import trueform_iscsi_targetextent.lun0 1
```
