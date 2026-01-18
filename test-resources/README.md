# Trueform Provider Test Suite

This directory contains Terraform configurations to test all resource types in the Trueform provider.

## Directory Structure

```
test-resources/
├── README.md           # This file
├── create/             # Creates one of each resource type
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   └── terraform.tfvars.example
├── modify/             # Modifies each resource (tests update operations)
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   └── terraform.tfvars.example
└── import/             # Imports all resources (tests import functionality)
    ├── main.tf
    ├── variables.tf
    ├── imports.tf
    └── terraform.tfvars
```

## Resources Tested

| Resource Type | Create Test | Modify Test | Import Test |
|--------------|-------------|-------------|-------------|
| `trueform_pool` | Creates test pool | N/A | Imports by ID |
| `trueform_dataset` | Creates dataset with lz4 compression | Changes to gzip, increases quota | Imports by path |
| `trueform_snapshot` | Creates snapshot | Creates new snapshot (immutable) | Imports by full path |
| `trueform_share_smb` | Creates SMB share | Enables guest access, read-only | Imports by ID |
| `trueform_share_nfs` | Creates NFS share with 1 network | Adds networks, sets read-only | Imports by ID |
| `trueform_iscsi_portal` | Creates iSCSI portal | Updates comment | Imports by ID |
| `trueform_iscsi_initiator` | Creates initiator with 1 IQN | Adds second IQN | Imports by ID |
| `trueform_iscsi_target` | Creates iSCSI target | Updates alias | Imports by ID |
| `trueform_iscsi_extent` | Creates 10MB file extent | Increases to 200MB | Imports by ID |
| `trueform_iscsi_targetextent` | Maps target to extent, LUN 0 | Changes to LUN 1 | Imports by ID |
| `trueform_user` | Creates test user | Enables sudo, updates email | Imports by ID |
| `trueform_cronjob` | Creates disabled daily cronjob | Enables, changes to hourly | Imports by ID |
| `trueform_static_route` | Creates static route | Updates description | Imports by ID |

## Prerequisites

1. A running TrueNAS Scale 25.04+ instance
2. An API key with appropriate permissions
3. An existing storage pool
4. The Trueform provider installed (via dev_overrides or registry)

## Usage

### Step 1: Configure Variables

```bash
cd create/
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your TrueNAS connection details
```

### Step 2: Create Resources

```bash
cd create/
terraform init
terraform plan
terraform apply
```

### Step 3: Test Modifications

```bash
cd ../modify/
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with matching values

# Copy state from create directory
cp ../create/terraform.tfstate .

terraform plan    # Should show updates, not creates
terraform apply   # Apply modifications
```

### Step 4: Test Imports (Alternative to Step 3)

This tests the provider's ability to import existing resources into Terraform state.

```bash
cd ../import/
# Edit terraform.tfvars with your TrueNAS connection details

# Get resource IDs from create directory
cd ../create
terraform show | grep -E "^\s+id\s+="
# Note all the IDs and update import/terraform.tfvars

cd ../import
terraform plan    # Should show imports, not creates
terraform apply   # Import all resources
terraform plan    # Should show NO changes (validates import worked correctly)
```

### Step 5: Clean Up

```bash
terraform destroy
```

## Variables Reference

### Required Variables

| Variable | Description |
|----------|-------------|
| `truenas_host` | TrueNAS host IP or hostname |
| `truenas_api_key` | API key for authentication |
| `pool_name` | Name of existing pool (e.g., "tank") |
| `base_path` | Base path for resources (e.g., "/mnt/tank") |
| `nfs_allowed_network` | CIDR for NFS access (create only) |
| `nfs_allowed_networks` | List of CIDRs for NFS (modify only) |
| `static_route_gateway` | Gateway IP for static route |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `truenas_verify_ssl` | `false` | Verify SSL certificates |
| `test_prefix` | `"tftest"` | Prefix for resource names |
| `iscsi_listen_ip` | `"0.0.0.0"` | iSCSI portal listen address |
| `static_route_destination` | `"10.99.99.0/24"` | Test route destination |
| `test_user_password` | varies | Password for test user |

## Notes

- The test prefix (`tftest` by default) is used for all resource names to avoid conflicts
- The cronjob is disabled by default in the create configuration for safety
- Snapshots are immutable, so the modify configuration creates a new snapshot
- Remember to destroy resources when done testing to clean up

## Troubleshooting

### "Resource already exists" errors
Ensure you're using the same `test_prefix` and that no conflicting resources exist.

### State mismatch between create and modify
Copy `terraform.tfstate` from `create/` to `modify/` before running apply in modify.

### Connection timeouts
Check that the TrueNAS host is reachable and the API is enabled.
