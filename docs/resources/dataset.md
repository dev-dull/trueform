---
page_title: "trueform_dataset Resource - Trueform"
subcategory: "Storage"
description: |-
  Manages a ZFS dataset on TrueNAS.
---

# trueform_dataset (Resource)

Manages a ZFS dataset on TrueNAS Scale. Datasets are the primary way to organize data on ZFS pools.

## Example Usage

### Basic Dataset

```hcl
resource "trueform_dataset" "media" {
  pool = "tank"
  name = "media"
}
```

### Dataset with Options

```hcl
resource "trueform_dataset" "documents" {
  pool        = "tank"
  name        = "documents"
  compression = "LZ4"
  atime       = "OFF"
  quota       = 1099511627776  # 1TB
  comments    = "Document storage"
}
```

### Nested Dataset

```hcl
resource "trueform_dataset" "photos" {
  pool = "tank"
  name = "media/photos"

  depends_on = [trueform_dataset.media]
}
```

## Schema

### Required

- `pool` (String) Name of the pool to create the dataset in.
- `name` (String) Name of the dataset. Use `/` for nested datasets (e.g., `parent/child`).

### Optional

- `atime` (String) Access time updates. Values: `ON`, `OFF`. Defaults to `ON`.
- `casesensitivity` (String) Case sensitivity. Values: `SENSITIVE`, `INSENSITIVE`, `MIXED`. Cannot be changed after creation.
- `comments` (String) Comments/description for the dataset.
- `compression` (String) Compression algorithm. Values: `OFF`, `LZ4`, `GZIP`, `ZSTD`, `ZLE`, `LZJB`. Defaults to `LZ4`.
- `copies` (Number) Number of data copies. Defaults to `1`.
- `deduplication` (String) Deduplication setting. Values: `ON`, `OFF`, `VERIFY`. Defaults to `OFF`.
- `quota` (Number) Quota in bytes. Must be >= 1GB or omitted.
- `readonly` (String) Read-only mode. Values: `ON`, `OFF`. Defaults to `OFF`.
- `recordsize` (String) Record size (e.g., `128K`).
- `share_type` (String) Share type preset. Values: `GENERIC`, `SMB`. Defaults to `GENERIC`.
- `snapdir` (String) Snapshot directory visibility. Values: `VISIBLE`, `HIDDEN`. Defaults to `HIDDEN`.
- `type` (String) Dataset type. Values: `FILESYSTEM`, `VOLUME`. Defaults to `FILESYSTEM`.

### Read-Only

- `aclmode` (String) ACL mode.
- `acltype` (String) ACL type.
- `available` (Number) Available space in bytes.
- `encrypted` (Boolean) Whether the dataset is encrypted.
- `encryption_root` (String) Encryption root dataset.
- `id` (String) Dataset identifier (pool/name format).
- `key_loaded` (Boolean) Whether encryption key is loaded.
- `managed_by` (String) Management source.
- `mountpoint` (String) Mount point path.
- `used` (Number) Used space in bytes.

## Import

Datasets can be imported using the pool/name format:

```shell
terraform import trueform_dataset.media tank/media
```
