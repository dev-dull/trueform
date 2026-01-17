---
page_title: "trueform_dataset Data Source - Trueform"
subcategory: "Storage"
description: |-
  Retrieves information about an existing ZFS dataset on TrueNAS.
---

# trueform_dataset (Data Source)

Retrieves information about an existing ZFS dataset on TrueNAS Scale.

## Example Usage

```hcl
data "trueform_dataset" "media" {
  id = "tank/media"
}

output "dataset_compression" {
  value = data.trueform_dataset.media.compression
}

output "dataset_used" {
  value = "${data.trueform_dataset.media.used / 1073741824} GB used"
}
```

## Schema

### Required

- `id` (String) Dataset identifier in pool/name format.

### Read-Only

- `aclmode` (String) ACL mode.
- `acltype` (String) ACL type.
- `atime` (String) Access time setting.
- `available` (Number) Available space in bytes.
- `casesensitivity` (String) Case sensitivity setting.
- `comments` (String) Dataset comments.
- `compression` (String) Compression algorithm.
- `deduplication` (String) Deduplication setting.
- `encrypted` (Boolean) Whether the dataset is encrypted.
- `mountpoint` (String) Mount point path.
- `name` (String) Dataset name.
- `pool` (String) Pool name.
- `quota` (Number) Quota in bytes.
- `readonly` (String) Read-only setting.
- `recordsize` (String) Record size.
- `type` (String) Dataset type.
- `used` (Number) Used space in bytes.
