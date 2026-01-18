# =============================================================================
# TrueNAS Connection Configuration
# =============================================================================
# Update these values to match your TrueNAS system

truenas_host       = "your-truenas-host"
truenas_api_key    = "your-api-key"
truenas_verify_ssl = false

# =============================================================================
# Resource Configuration
# =============================================================================
# These should match the values used in the 'create' directory

pool_name            = "testpool"
pool_disks           = ["xvdb", "xvdc", "xvde", "xvdf"]
nfs_allowed_network  = "192.168.1.0/24"
static_route_gateway = "192.168.1.1"

# =============================================================================
# Import IDs
# =============================================================================
# Update these values with the IDs from the resources created by the 'create'
# configuration. Run 'terraform show' in the 'create' directory to get the IDs.
#
# Example workflow:
# 1. cd ../create && terraform apply
# 2. terraform show | grep -E "^  id|^  name" (to get the IDs)
# 3. Update the values below
# 4. cd ../import && terraform apply

pool_id               = 1
dataset_id            = "testpool/tftest_dataset"
snapshot_id           = "testpool/tftest_dataset@tftest_snapshot"
smb_share_id          = 1
nfs_share_id          = 1
iscsi_portal_id       = 1
iscsi_initiator_id    = 1
iscsi_target_id       = 1
iscsi_extent_id       = 1
iscsi_targetextent_id = 1
user_id               = 1
cronjob_id            = 1
static_route_id       = 1
