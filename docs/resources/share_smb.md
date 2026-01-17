---
page_title: "trueform_share_smb Resource - Trueform"
subcategory: "Sharing"
description: |-
  Manages an SMB/CIFS share on TrueNAS.
---

# trueform_share_smb (Resource)

Manages an SMB (Server Message Block) / CIFS share on TrueNAS Scale. SMB shares provide Windows-compatible file sharing.

## Example Usage

### Basic SMB Share

```hcl
resource "trueform_share_smb" "documents" {
  name    = "documents"
  path    = "/mnt/tank/documents"
  enabled = true
}
```

### SMB Share with Options

```hcl
resource "trueform_share_smb" "public" {
  name      = "public"
  path      = "/mnt/tank/public"
  enabled   = true
  browsable = true
  guestok   = true
  ro        = true
  comment   = "Public read-only share"
}
```

## Schema

### Required

- `name` (String) Share name (visible to clients).
- `path` (String) Filesystem path to share.

### Optional

- `abe` (Boolean) Access Based Enumeration - hide files/folders users cannot access. Defaults to `false`.
- `acl` (Boolean) Enable ACL support. Defaults to `true`.
- `browsable` (Boolean) Share is visible in network browse lists. Defaults to `true`.
- `comment` (String) Share description/comment.
- `durablehandle` (Boolean) Enable durable handles. Defaults to `true`.
- `enabled` (Boolean) Enable the share. Defaults to `true`.
- `fsrvp` (Boolean) Enable File Server Remote VSS Protocol. Defaults to `false`.
- `guestok` (Boolean) Allow guest access. Defaults to `false`.
- `home` (Boolean) Use as home share. Defaults to `false`.
- `purpose` (String) Purpose preset. Values: `DEFAULT_SHARE`, `TIMEMACHINE_SHARE`, etc. Defaults to `DEFAULT_SHARE`.
- `recyclebin` (Boolean) Enable recycle bin. Defaults to `false`.
- `ro` (Boolean) Read-only share. Defaults to `false`.
- `shadowcopy` (Boolean) Enable shadow copies (Previous Versions). Defaults to `true`.
- `streams` (Boolean) Enable alternate data streams. Defaults to `true`.
- `timemachine` (Boolean) Enable Time Machine support. Defaults to `false`.

### Read-Only

- `audit_logging` (Boolean) Whether audit logging is enabled.
- `id` (Number) Share identifier.
- `locked` (Boolean) Whether the share is locked.

## Import

SMB shares can be imported using the share ID:

```shell
terraform import trueform_share_smb.documents 1
```

## Notes

Some attributes (`ro`, `guestok`, `recyclebin`, `abe`, `browsable`) cannot be updated after creation in TrueNAS Scale 25. To change these values, you must recreate the share.
