# =============================================================================
# Import Blocks - Import existing TrueNAS resources into Terraform state
# =============================================================================
# These import blocks allow Terraform to adopt existing resources that were
# created by the 'create' configuration.
#
# After running 'terraform apply' in the 'create' directory, note the IDs of
# all created resources and update the IDs below.
#
# Then run 'terraform apply' here to import all resources.
# =============================================================================

# Pool
import {
  to = trueform_pool.test
  id = "1"
}

# Dataset
import {
  to = trueform_dataset.test
  id = "testpool/tftest_dataset"
}

# Snapshot
import {
  to = trueform_snapshot.test
  id = "testpool/tftest_dataset@tftest_snapshot"
}

# SMB Share
import {
  to = trueform_share_smb.test
  id = "1"
}

# NFS Share
import {
  to = trueform_share_nfs.test
  id = "1"
}

# iSCSI Portal
import {
  to = trueform_iscsi_portal.test
  id = "1"
}

# iSCSI Initiator
import {
  to = trueform_iscsi_initiator.test
  id = "1"
}

# iSCSI Target
import {
  to = trueform_iscsi_target.test
  id = "1"
}

# iSCSI Extent
import {
  to = trueform_iscsi_extent.test
  id = "1"
}

# iSCSI Target Extent Mapping
import {
  to = trueform_iscsi_targetextent.test
  id = "1"
}

# User
import {
  to = trueform_user.test
  id = "1"
}

# Cronjob
import {
  to = trueform_cronjob.test
  id = "1"
}

# Static Route
import {
  to = trueform_static_route.test
  id = "1"
}
