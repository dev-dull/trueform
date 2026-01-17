---
page_title: "trueform_iscsi_target Resource - Trueform"
subcategory: "iSCSI"
description: |-
  Manages an iSCSI target on TrueNAS.
---

# trueform_iscsi_target (Resource)

Manages an iSCSI target on TrueNAS Scale. Targets are the named endpoints that initiators connect to.

## Example Usage

### Basic Target

```hcl
resource "trueform_iscsi_target" "storage" {
  name  = "storage"
  alias = "Storage Target"
}
```

### Target with Portal Group

```hcl
resource "trueform_iscsi_target" "data" {
  name  = "data"
  alias = "Data Target"
  mode  = "ISCSI"

  groups = [
    {
      portal    = trueform_iscsi_portal.default.id
      initiator = trueform_iscsi_initiator.trusted.id
    }
  ]
}
```

## Schema

### Required

- `name` (String) Target name (will be prefixed with system IQN). Cannot be changed after creation.

### Optional

- `alias` (String) Target alias/description.
- `groups` (List of Object) Portal groups for the target.
  - `portal` (Number, Required) Portal ID.
  - `initiator` (Number, Optional) Initiator group ID.
  - `auth` (Number, Optional) Auth credential group ID.
  - `authmethod` (String, Optional) Authentication method. Values: `NONE`, `CHAP`, `CHAP_MUTUAL`.
- `mode` (String) Target mode. Values: `ISCSI`, `FC`, `BOTH`. Defaults to `ISCSI`.

### Read-Only

- `id` (Number) Target identifier.

## Import

iSCSI targets can be imported using the target ID:

```shell
terraform import trueform_iscsi_target.storage 1
```
