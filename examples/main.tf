terraform {
  required_providers {
    trueform = {
      source = "registry.terraform.io/trueform/trueform"
    }
  }
}

# Configure the TrueNAS provider
provider "trueform" {
  host       = var.truenas_host
  api_key    = var.truenas_api_key
  verify_ssl = var.truenas_verify_ssl
}

variable "truenas_host" {
  description = "TrueNAS host address"
  type        = string
}

variable "truenas_api_key" {
  description = "TrueNAS API key"
  type        = string
  sensitive   = true
}

variable "truenas_verify_ssl" {
  description = "Verify SSL certificates"
  type        = bool
  default     = true
}

# Data source example: Look up an existing pool
data "trueform_pool" "main" {
  name = "tank"
}

# Create a dataset
resource "trueform_dataset" "media" {
  pool        = data.trueform_pool.main.name
  name        = "media"
  compression = "lz4"
  quota       = 1099511627776  # 1TB

  comments = "Media storage dataset"
}

# Create an SMB share
resource "trueform_share_smb" "media" {
  path    = "/mnt/${data.trueform_pool.main.name}/media"
  name    = "media"
  comment = "Media share"
  enabled = true

  browsable = true
  guestok   = false
  ro        = false

  depends_on = [trueform_dataset.media]
}

# Create an NFS share
resource "trueform_share_nfs" "media" {
  path    = "/mnt/${data.trueform_pool.main.name}/media"
  enabled = true

  networks = ["192.168.1.0/24"]

  maproot_user  = "root"
  maproot_group = "wheel"

  depends_on = [trueform_dataset.media]
}

# Create a user
resource "trueform_user" "media_user" {
  username  = "mediauser"
  full_name = "Media User"
  password  = "changeme123"

  home  = "/mnt/${data.trueform_pool.main.name}/media"
  shell = "/usr/bin/zsh"

  smb = true
}

# Create a cron job
resource "trueform_cronjob" "backup_script" {
  user        = "root"
  command     = "/usr/local/bin/backup.sh"
  description = "Daily backup script"
  enabled     = true

  schedule {
    minute = "0"
    hour   = "2"
    dom    = "*"
    month  = "*"
    dow    = "*"
  }
}

# Create a VM
resource "trueform_vm" "ubuntu" {
  name        = "ubuntu-server"
  description = "Ubuntu Server VM"

  vcpus   = 2
  cores   = 1
  threads = 1
  memory  = 4096

  bootloader = "UEFI"
  autostart  = true
}

# Add a disk to the VM
resource "trueform_vm_device" "ubuntu_disk" {
  vm    = trueform_vm.ubuntu.id
  dtype = "DISK"
  order = 1001

  disk_path = "zvol/${data.trueform_pool.main.name}/vms/ubuntu-boot"
  disk_type = "VIRTIO"
}

# Add a NIC to the VM
resource "trueform_vm_device" "ubuntu_nic" {
  vm    = trueform_vm.ubuntu.id
  dtype = "NIC"
  order = 1002

  nic_type   = "VIRTIO"
  nic_attach = "br0"
}

# iSCSI configuration example
resource "trueform_iscsi_portal" "default" {
  comment = "Default iSCSI portal"

  listen {
    ip   = "0.0.0.0"
    port = 3260
  }
}

resource "trueform_iscsi_initiator" "all" {
  comment = "Allow all initiators"
  # Empty initiators list allows all
}

resource "trueform_iscsi_target" "storage" {
  name  = "iqn.2024.storage"
  alias = "Storage Target"
  mode  = "ISCSI"

  groups {
    portal    = trueform_iscsi_portal.default.id
    initiator = trueform_iscsi_initiator.all.id
  }
}

resource "trueform_iscsi_extent" "lun0" {
  name = "storage-lun0"
  type = "DISK"
  disk = "zvol/${data.trueform_pool.main.name}/iscsi/lun0"

  blocksize = 512
  rpm       = "SSD"
}

resource "trueform_iscsi_targetextent" "lun0_mapping" {
  target = trueform_iscsi_target.storage.id
  extent = trueform_iscsi_extent.lun0.id
  lunid  = 0
}

# Outputs
output "pool_info" {
  value = {
    name      = data.trueform_pool.main.name
    status    = data.trueform_pool.main.status
    size      = data.trueform_pool.main.size
    free      = data.trueform_pool.main.free
  }
}

output "smb_share_name" {
  value = trueform_share_smb.media.name
}

output "vm_id" {
  value = trueform_vm.ubuntu.id
}
