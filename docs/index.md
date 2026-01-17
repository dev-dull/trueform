---
page_title: "Trueform Provider"
description: |-
  The Trueform provider enables Terraform to manage TrueNAS Scale resources.
---

# Trueform Provider

The Trueform provider enables Terraform to manage [TrueNAS Scale](https://www.truenas.com/truenas-scale/) resources using the WebSocket JSON-RPC API.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [TrueNAS Scale](https://www.truenas.com/truenas-scale/) >= 25.04
- A TrueNAS API key with appropriate permissions

## Installation

The provider is available from the [Terraform Registry](https://registry.terraform.io/providers/trueform/trueform/latest).

```hcl
terraform {
  required_providers {
    trueform = {
      source  = "trueform/trueform"
      version = "~> 0.1"
    }
  }
}
```

## Authentication

The provider requires an API key for authentication. You can create an API key in the TrueNAS web UI under **Credentials > API Keys**.

### Configuration

```hcl
provider "trueform" {
  host       = "192.168.1.100"
  api_key    = var.truenas_api_key
  verify_ssl = false
}
```

### Environment Variables

Alternatively, configure the provider using environment variables:

```bash
export TRUENAS_HOST="192.168.1.100"
export TRUENAS_API_KEY="1-your-api-key-here"
export TRUENAS_VERIFY_SSL="false"
```

```hcl
provider "trueform" {
  # Configuration from environment variables
}
```

## Example Usage

```hcl
terraform {
  required_providers {
    trueform = {
      source  = "trueform/trueform"
      version = "~> 0.1"
    }
  }
}

provider "trueform" {
  host       = "192.168.1.100"
  api_key    = var.truenas_api_key
  verify_ssl = false
}

# Look up an existing pool
data "trueform_pool" "main" {
  name = "tank"
}

# Create a dataset
resource "trueform_dataset" "media" {
  pool        = data.trueform_pool.main.name
  name        = "media"
  compression = "LZ4"
  comments    = "Media storage"
}

# Create an SMB share
resource "trueform_share_smb" "media" {
  name    = "media"
  path    = "/mnt/tank/media"
  enabled = true

  depends_on = [trueform_dataset.media]
}
```

## Schema

### Required

- `host` (String) TrueNAS host address (IP or hostname).
- `api_key` (String, Sensitive) TrueNAS API key for authentication.

### Optional

- `verify_ssl` (Boolean) Whether to verify SSL certificates. Defaults to `true`.
