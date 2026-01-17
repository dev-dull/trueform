---
page_title: "trueform_user Resource - Trueform"
subcategory: "Accounts"
description: |-
  Manages a local user account on TrueNAS.
---

# trueform_user (Resource)

Manages a local user account on TrueNAS Scale.

## Example Usage

### Basic User

```hcl
resource "trueform_user" "john" {
  username  = "john"
  full_name = "John Doe"
  password  = var.user_password
}
```

### User with SMB Access

```hcl
resource "trueform_user" "mediauser" {
  username  = "mediauser"
  full_name = "Media User"
  password  = var.media_password
  email     = "media@example.com"

  shell = "/usr/sbin/nologin"
  smb   = true
}
```

### User with Home Directory

```hcl
resource "trueform_user" "developer" {
  username     = "developer"
  full_name    = "Developer Account"
  password     = var.dev_password

  home         = "/mnt/tank/homes/developer"
  home_create  = true
  home_mode    = "755"
  shell        = "/bin/bash"
  group_create = true
}
```

## Schema

### Required

- `username` (String) Username for the account.
- `password` (String, Sensitive) User password.

### Optional

- `email` (String) User email address.
- `full_name` (String) User's full name.
- `group` (Number) Primary group ID.
- `group_create` (Boolean) Create a new primary group for the user. Defaults to `true`.
- `groups` (List of Number) Additional group IDs.
- `home` (String) Home directory path. Defaults to `/var/empty`.
- `home_create` (Boolean) Create home directory if it doesn't exist. Defaults to `false`.
- `home_mode` (String) Home directory permissions (e.g., `700`). Defaults to `700`.
- `locked` (Boolean) Lock the account. Defaults to `false`.
- `password_disabled` (Boolean) Disable password authentication. Defaults to `false`.
- `shell` (String) Login shell. Defaults to `/usr/sbin/nologin`.
- `smb` (Boolean) Enable SMB authentication. Defaults to `true`.
- `sudo` (Boolean) Grant sudo privileges. Defaults to `false`.
- `sudo_nopasswd` (Boolean) Allow sudo without password. Defaults to `false`.

### Read-Only

- `builtin` (Boolean) Whether this is a built-in system account.
- `id` (Number) User identifier.
- `uid` (Number) Unix user ID.

## Import

Users can be imported using the user ID:

```shell
terraform import trueform_user.john 1001
```
