---
page_title: "trueform_user Data Source - Trueform"
subcategory: "Accounts"
description: |-
  Retrieves information about an existing user on TrueNAS.
---

# trueform_user (Data Source)

Retrieves information about an existing local user account on TrueNAS Scale.

## Example Usage

```hcl
data "trueform_user" "admin" {
  username = "admin"
}

output "admin_uid" {
  value = data.trueform_user.admin.uid
}
```

### Using in NFS Share

```hcl
data "trueform_user" "nfs_user" {
  username = "nfsuser"
}

resource "trueform_share_nfs" "data" {
  path         = "/mnt/tank/data"
  maproot_user = data.trueform_user.nfs_user.username
}
```

## Schema

### Required

- `username` (String) Username to look up.

### Read-Only

- `builtin` (Boolean) Whether this is a built-in system account.
- `email` (String) User email address.
- `full_name` (String) User's full name.
- `group` (Number) Primary group ID.
- `home` (String) Home directory path.
- `id` (Number) User identifier.
- `locked` (Boolean) Whether the account is locked.
- `shell` (String) Login shell.
- `smb` (Boolean) SMB authentication enabled.
- `uid` (Number) Unix user ID.
