---
page_title: "trueform_iscsi_initiator Resource - Trueform"
subcategory: "iSCSI"
description: |-
  Manages an iSCSI initiator group on TrueNAS.
---

# trueform_iscsi_initiator (Resource)

Manages an iSCSI initiator group on TrueNAS Scale. Initiator groups define which clients are allowed to connect to iSCSI targets.

## Example Usage

### Allow All Initiators

```hcl
resource "trueform_iscsi_initiator" "all" {
  comment = "Allow all initiators"
  # Empty initiators list allows all
}
```

### Specific Initiators

```hcl
resource "trueform_iscsi_initiator" "trusted" {
  comment = "Trusted servers"

  initiators = [
    "iqn.2024-01.com.example:server1",
    "iqn.2024-01.com.example:server2",
  ]
}
```

## Schema

### Optional

- `comment` (String) Initiator group description.
- `initiators` (List of String) List of allowed initiator IQNs. Empty list allows all initiators.

### Read-Only

- `id` (Number) Initiator group identifier.

## Import

iSCSI initiator groups can be imported using the group ID:

```shell
terraform import trueform_iscsi_initiator.trusted 1
```
