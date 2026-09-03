<p align="center">
  <img src="logo.svg" alt="TrueForm Logo" width="200">
</p>

# TrueForm

A Terraform provider for managing TrueNAS Scale 25.04+ and TrueNAS 26 (beta) resources.

## About

There didn't appear to be any published Terraform providers for TrueNAS Scale for versions 25 and greater (due to the recent change in how the API works), so I decided to burn my Claude Code trial by generating one. It has been mostly written by Claude Code by pointing it at the [TrueNAS 25 documentation](https://www.truenas.com/docs/scale/25.04/api/).

The name TrueForm is taking "form" from Terraform, and "true" from TrueNAS.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (for building from source)
- TrueNAS Scale 25.04+, or TrueNAS 26 (beta) — the provider detects the server version at runtime and uses the correct API for each (verified end-to-end against 25.10.4 and 26.0.0-BETA.2)

## Installation

### Building from Source

```bash
git clone https://github.com/trueform/terraform-provider-trueform.git
cd trueform
go build -o terraform-provider-trueform
```

### Local Installation

After building, move the binary to your Terraform plugins directory:

```bash
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/trueform/trueform/0.1.0/darwin_arm64/
mv terraform-provider-trueform ~/.terraform.d/plugins/registry.terraform.io/trueform/trueform/0.1.0/darwin_arm64/
```

Adjust the path for your OS/architecture (e.g., `linux_amd64`, `darwin_amd64`).

## Authentication

The provider uses API key authentication. To create an API key in TrueNAS:

1. Log into your TrueNAS web interface
2. Click on your username in the top-right corner
3. Select **API Keys**
4. Click **Add** and give your key a name
5. Copy the generated key (you won't be able to see it again)

## Configuration

```hcl
terraform {
  required_providers {
    trueform = {
      source = "registry.terraform.io/trueform/trueform"
    }
  }
}

provider "trueform" {
  host       = "truenas.local"      # Your TrueNAS hostname or IP
  api_key    = var.truenas_api_key  # Your API key
  verify_ssl = true                 # Set to false for self-signed certs
}
```

You can also use environment variables:

```bash
export TRUENAS_HOST="truenas.local"
export TRUENAS_API_KEY="your-api-key"
export TRUENAS_VERIFY_SSL="true"
```

## Available Resources

| Resource | Description |
|----------|-------------|
| `trueform_pool` | Manage ZFS storage pools |
| `trueform_dataset` | Manage ZFS datasets |
| `trueform_snapshot` | Manage ZFS snapshots |
| `trueform_share_smb` | Manage SMB/CIFS shares |
| `trueform_share_nfs` | Manage NFS exports |
| `trueform_user` | Manage local users |
| `trueform_vm` | Manage virtual machines |
| `trueform_vm_device` | Manage VM devices (disks, NICs, etc.) |
| `trueform_app` | Manage applications |
| `trueform_service_docker` | Manage Docker/Apps service configuration |
| `trueform_cronjob` | Manage scheduled tasks |
| `trueform_certificate` | Manage SSL/TLS certificates |
| `trueform_static_route` | Manage network routes |
| `trueform_iscsi_portal` | Manage iSCSI portals |
| `trueform_iscsi_target` | Manage iSCSI targets |
| `trueform_iscsi_extent` | Manage iSCSI extents (LUNs) |
| `trueform_iscsi_initiator` | Manage iSCSI initiator groups |
| `trueform_iscsi_targetextent` | Manage iSCSI target-to-extent mappings |

## Available Data Sources

| Data Source | Description |
|-------------|-------------|
| `trueform_pool` | Query existing pools |
| `trueform_dataset` | Query existing datasets |
| `trueform_user` | Query existing users |
| `trueform_vm` | Query existing VMs |

## Usage Examples

### Create a Dataset

```hcl
resource "trueform_dataset" "media" {
  pool        = "tank"
  name        = "media"
  compression = "LZ4"
  quota       = 1099511627776  # 1TB in bytes
  comments    = "Media storage"
}
```

### Create an SMB Share

```hcl
resource "trueform_share_smb" "media" {
  path      = "/mnt/tank/media"
  name      = "media"
  comment   = "Media share"
  enabled   = true
  browsable = true
  guestok   = false
}
```

### Create a Virtual Machine

```hcl
resource "trueform_vm" "ubuntu" {
  name        = "ubuntu-server"
  description = "Ubuntu Server VM"
  vcpus       = 2
  cores       = 1
  threads     = 1
  memory      = 4096  # MB
  bootloader  = "UEFI"
  autostart   = true
}

resource "trueform_vm_device" "ubuntu_disk" {
  vm        = trueform_vm.ubuntu.id
  dtype     = "DISK"
  disk_path = "zvol/tank/vms/ubuntu-boot"
  disk_type = "VIRTIO"
}

resource "trueform_vm_device" "ubuntu_nic" {
  vm         = trueform_vm.ubuntu.id
  dtype      = "NIC"
  nic_type   = "VIRTIO"
  nic_attach = "br0"
}
```

### Query Existing Resources

```hcl
data "trueform_pool" "main" {
  name = "tank"
}

output "pool_free_space" {
  value = data.trueform_pool.main.free
}
```

See the [examples](./examples/) directory for more complete examples.

## Development

### Building

```bash
go build -o terraform-provider-trueform
```

### Running Tests

```bash
go test ./...
```

### Running Acceptance Tests

Acceptance tests run against a real TrueNAS instance:

```bash
export TRUENAS_HOST="your-truenas-host"
export TRUENAS_API_KEY="your-api-key"
TF_ACC=1 go test ./... -v
```

### Releasing

This project uses [Semantic Versioning](https://semver.org/):

- **MAJOR** — breaking changes to resource schemas or provider configuration
- **MINOR** — new resources, data sources, or features
- **PATCH** — bug fixes, documentation updates, lint fixes

To cut a release:

1. Push your changes to `main`.
2. Wait for CI (lint, build, and tests) to pass.
3. Tag the passing commit:
   ```bash
   git tag v0.X.Y
   git push origin v0.X.Y
   ```

The Terraform Registry picks up new versions via GitHub tag-creation webhooks. Do not force-push tags — if a tag needs to be revised, create a new patch version instead.

## Technical Details

This provider communicates with TrueNAS using the WebSocket JSON-RPC 2.0 API introduced in TrueNAS Scale 25.04. The connection flow is:

1. Establish WebSocket connection to `wss://<host>/api/current`
2. Authenticate using `auth.login_ex` with the `API_KEY_PLAIN` mechanism (falls back to `auth.login_with_api_key` for older TrueNAS versions)
3. Execute JSON-RPC calls for resource operations

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

This project is provided as-is without any specific license. Use at your own risk.

## Disclaimer

This provider is not officially affiliated with or endorsed by iXsystems or the TrueNAS project. Use at your own risk and always test in a non-production environment first.
