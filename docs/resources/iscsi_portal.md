---
page_title: "trueform_iscsi_portal Resource - Trueform"
subcategory: "iSCSI"
description: |-
  Manages an iSCSI portal on TrueNAS.
---

# trueform_iscsi_portal (Resource)

Manages an iSCSI portal on TrueNAS Scale. Portals define the network endpoints where iSCSI targets listen for connections.

## Example Usage

### Basic Portal

```hcl
resource "trueform_iscsi_portal" "default" {
  comment = "Default iSCSI portal"

  listen = [
    {
      ip   = "0.0.0.0"
      port = 3260
    }
  ]
}
```

### Portal with Specific IP

```hcl
resource "trueform_iscsi_portal" "storage_network" {
  comment = "Storage network portal"

  listen = [
    {
      ip   = "10.0.0.100"
      port = 3260
    }
  ]
}
```

### Portal with Multiple Listen Addresses

```hcl
resource "trueform_iscsi_portal" "multi" {
  comment = "Multi-homed portal"

  listen = [
    {
      ip   = "192.168.1.100"
      port = 3260
    },
    {
      ip   = "10.0.0.100"
      port = 3260
    }
  ]
}
```

## Schema

### Required

- `listen` (List of Object) Listen addresses for the portal.
  - `ip` (String) IP address to listen on. Use `0.0.0.0` for all interfaces.
  - `port` (Number, Optional) TCP port. Defaults to `3260`.

### Optional

- `comment` (String) Portal description.
- `discovery_authgroup` (Number) Discovery authentication group ID.
- `discovery_authmethod` (String) Discovery authentication method. Values: `NONE`, `CHAP`, `CHAP_MUTUAL`.

### Read-Only

- `id` (Number) Portal identifier.

## Import

iSCSI portals can be imported using the portal ID:

```shell
terraform import trueform_iscsi_portal.default 1
```
