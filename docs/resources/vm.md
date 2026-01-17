---
page_title: "trueform_vm Resource - Trueform"
subcategory: "Virtualization"
description: |-
  Manages a virtual machine on TrueNAS.
---

# trueform_vm (Resource)

Manages a virtual machine on TrueNAS Scale.

## Example Usage

### Basic VM

```hcl
resource "trueform_vm" "ubuntu" {
  name        = "ubuntu-server"
  description = "Ubuntu Server VM"

  vcpus   = 2
  cores   = 1
  threads = 1
  memory  = 4096

  bootloader = "UEFI"
  autostart  = false
}
```

### VM with All Options

```hcl
resource "trueform_vm" "windows" {
  name        = "windows-server"
  description = "Windows Server VM"

  vcpus    = 4
  cores    = 2
  threads  = 2
  memory   = 8192
  min_memory = 4096

  bootloader = "UEFI"
  autostart  = true
  time       = "LOCAL"
}
```

## Schema

### Required

- `name` (String) VM name.

### Optional

- `autostart` (Boolean) Start VM automatically on boot. Defaults to `false`.
- `bootloader` (String) Bootloader type. Values: `UEFI`, `UEFI_CSM`. Defaults to `UEFI`.
- `cores` (Number) CPU cores per socket. Defaults to `1`.
- `description` (String) VM description.
- `memory` (Number) Memory in MB.
- `min_memory` (Number) Minimum memory for ballooning in MB.
- `threads` (Number) CPU threads per core. Defaults to `1`.
- `time` (String) VM clock type. Values: `LOCAL`, `UTC`. Defaults to `LOCAL`.
- `vcpus` (Number) Number of virtual CPUs. Defaults to `1`.

### Read-Only

- `id` (Number) VM identifier.
- `status` (String) VM status.

## Import

VMs can be imported using the VM ID:

```shell
terraform import trueform_vm.ubuntu 1
```
