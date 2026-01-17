---
page_title: "trueform_vm_device Resource - Trueform"
subcategory: "Virtualization"
description: |-
  Manages a virtual machine device on TrueNAS.
---

# trueform_vm_device (Resource)

Manages a virtual machine device on TrueNAS Scale. Devices include disks, NICs, CD-ROMs, and PCI passthrough.

## Example Usage

### Disk Device

```hcl
resource "trueform_vm_device" "disk" {
  vm    = trueform_vm.ubuntu.id
  dtype = "DISK"
  order = 1001

  disk_path = "/dev/zvol/tank/vms/ubuntu-boot"
  disk_type = "VIRTIO"
}
```

### NIC Device

```hcl
resource "trueform_vm_device" "nic" {
  vm    = trueform_vm.ubuntu.id
  dtype = "NIC"
  order = 1002

  nic_type   = "VIRTIO"
  nic_attach = "br0"
}
```

### CD-ROM Device

```hcl
resource "trueform_vm_device" "cdrom" {
  vm    = trueform_vm.ubuntu.id
  dtype = "CDROM"
  order = 1000

  cdrom_path = "/mnt/tank/iso/ubuntu-22.04.iso"
}
```

## Schema

### Required

- `dtype` (String) Device type. Values: `DISK`, `NIC`, `CDROM`, `PCI`, `DISPLAY`, `RAW`.
- `vm` (Number) VM ID to attach the device to.

### Optional

- `order` (Number) Boot order (lower numbers boot first).

### Disk Options (dtype = DISK)

- `disk_path` (String) Path to zvol or disk.
- `disk_type` (String) Disk interface type. Values: `VIRTIO`, `AHCI`, `SCSI`.
- `disk_sectorsize` (Number) Logical sector size.

### NIC Options (dtype = NIC)

- `nic_attach` (String) Network bridge to attach to.
- `nic_mac` (String) MAC address (auto-generated if not specified).
- `nic_type` (String) NIC type. Values: `VIRTIO`, `E1000`.

### CD-ROM Options (dtype = CDROM)

- `cdrom_path` (String) Path to ISO file.

### Read-Only

- `id` (Number) Device identifier.

## Import

VM devices can be imported using the device ID:

```shell
terraform import trueform_vm_device.disk 1
```
