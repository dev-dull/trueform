---
page_title: "trueform_share_nfs Resource - Trueform"
subcategory: "Sharing"
description: |-
  Manages an NFS export on TrueNAS.
---

# trueform_share_nfs (Resource)

Manages an NFS (Network File System) export on TrueNAS Scale. NFS shares provide Unix/Linux-compatible file sharing.

## Example Usage

### Basic NFS Share

```hcl
resource "trueform_share_nfs" "data" {
  path    = "/mnt/tank/data"
  enabled = true
}
```

### NFS Share with Network Restrictions

```hcl
resource "trueform_share_nfs" "internal" {
  path    = "/mnt/tank/internal"
  enabled = true

  networks = [
    "192.168.1.0/24",
    "10.0.0.0/8",
  ]
}
```

### NFS Share with User Mapping

```hcl
resource "trueform_share_nfs" "media" {
  path    = "/mnt/tank/media"
  enabled = true
  ro      = true

  networks      = ["192.168.1.0/24"]
  maproot_user  = "root"
  maproot_group = "wheel"
}
```

## Schema

### Required

- `path` (String) Filesystem path to export.

### Optional

- `aliases` (List of String) Path aliases for the export.
- `comment` (String) Export description/comment.
- `enabled` (Boolean) Enable the export. Defaults to `true`.
- `hosts` (List of String) List of allowed hostnames or IP addresses.
- `mapall_group` (String) Map all client groups to this group.
- `mapall_user` (String) Map all client users to this user.
- `maproot_group` (String) Map root group to this group.
- `maproot_user` (String) Map root user to this user.
- `networks` (List of String) List of allowed networks in CIDR notation.
- `ro` (Boolean) Read-only export. Defaults to `false`.
- `security` (List of String) Security flavors. Values: `sys`, `krb5`, `krb5i`, `krb5p`.

### Read-Only

- `id` (Number) Export identifier.
- `locked` (Boolean) Whether the export is locked.

## Import

NFS exports can be imported using the export ID:

```shell
terraform import trueform_share_nfs.data 1
```
