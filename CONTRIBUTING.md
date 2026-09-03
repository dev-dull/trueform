# Contributing

## Test VM Setup (Proxmox)

The integration tests in `test-resources/` require a TrueNAS SCALE VM with disposable disks. This guide covers setting up the test VM on Proxmox VE, taking a clean snapshot, and the snapshot-revert workflow for repeatable test runs.

### 1. Download TrueNAS SCALE ISO

Download TrueNAS SCALE 25.04 (or newer) from https://www.truenas.com/download-truenas-scale/ and upload it to a Proxmox storage backend that supports ISO images (e.g., `local` or an NFS ISO library).

### 2. Create the VM

Create a new VM in the Proxmox UI with the following specs:

| Setting | Value |
|---------|-------|
| CPU | 2 cores |
| RAM | 8 GB |
| Boot disk | 32 GB SCSI, **qcow2 format** |
| Test disks | 4x 512 MB SCSI disks, **qcow2 format** |
| Network | 1x virtio NIC on a bridged network with DHCP |
| OS | Attach the TrueNAS SCALE ISO |

The 4 test disks are used by the integration tests to create and destroy ZFS pools. They must be small (512 MB is sufficient) and separate from the boot disk.

**Important: use qcow2 disk format.** Proxmox snapshots require qcow2 when using NFS-backed storage (e.g., `truenas-vdisk`). Raw format disks do not support snapshots on NFS — the snapshot will fail with "snapshot feature is not available". If you accidentally created the VM with raw disks, you can convert them while the VM is stopped:

```
qm move_disk <vmid> scsi0 <storage> --format qcow2 --delete 1
```

Repeat for each disk (`scsi0` through `scsi4`). The conversions must run one at a time (Proxmox locks the VM config during each move). Local storage backends (LVM-thin, ZFS) support snapshots with any format.

**Disk naming:** Proxmox SCSI disks appear as `sd*` in the guest. The boot disk is typically `sda`, so the 4 test disks will be `sdb`, `sdc`, `sdd`, `sde`. Verify after install using the `disk.query` API (see step 5).

### 3. Install TrueNAS

Boot the VM from the ISO and follow the TrueNAS installer:

1. Select the 32 GB boot disk as the install target
2. Set the admin password
3. Reboot (remove the ISO from the CD drive first)

### 4. Initial Configuration

After the VM boots into TrueNAS:

1. Complete the setup wizard in the web UI
2. Navigate to **Credentials > API Keys > Add**
3. Create an API key and save it — you'll need it for `terraform.tfvars`

Store the VM's IP and API key in `test-resources/creds` (gitignored):

```
IP: <vm-ip>
API_KEY: <api-key>
```

### 5. Discover Disk Names

Query the TrueNAS API to confirm the disk identifiers:

```bash
curl -k -X POST "https://<HOST>/api/current" \
  -H "Authorization: Bearer <API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"msg":"method","method":"disk.query","params":[[],{"select":["name","size","type"]}]}'
```

You should see `sda` (boot) and `sdb`, `sdc`, `sdd`, `sde` (test disks). Update `pool_disks` in the `terraform.tfvars` files accordingly.

### 6. Take a Clean Snapshot

In the Proxmox UI, take a VM snapshot (e.g., named `clean-install`). This captures the fresh TrueNAS state before any test resources are created.

### Snapshot-Revert Workflow

Before each full test cycle (`create → import → modify → destroy`), revert the VM to the `clean-install` snapshot. This guarantees:

- No leftover pools, datasets, shares, or users from previous runs
- The API key remains valid (it was created before the snapshot)
- Disk identifiers stay consistent

To revert via the Proxmox API or CLI:

```bash
# CLI (on the Proxmox host)
qm snapshot <vmid> rollback clean-install

# API
curl -X POST "https://<proxmox-host>:8006/api2/json/nodes/<node>/qemu/<vmid>/snapshot/clean-install/rollback" \
  -H "Authorization: PVEAPIToken=<token>"
```

Or simply use the Proxmox UI: select the VM → Snapshots → `clean-install` → Rollback.

After reverting, wait for TrueNAS to boot (~30–60 seconds), then run the test cycle as described in `test-resources/README.md`.

### Testing Against Newer TrueNAS Releases

The point of this VM is to exercise the **trueform provider** against current TrueNAS
releases. TrueNAS updates themselves are trusted — iX tests them; validating TrueNAS
is not the goal here. So the workflow is simply to keep an up-to-date base snapshot
per TrueNAS version you support (a "golden") and run the provider test cycle against
each one.

Keep one golden per major version — e.g. a 25.x golden and a 26.x golden. To refresh
a golden and test the provider against the update, per version:

1. **Revert** the VM to that version's current golden snapshot (Proxmox rollback).
2. **Apply TrueNAS updates** in the TrueNAS UI/API and let it reboot. For a beta
   line (26, until 26.0 GA) deciding *whether* to move to the next BETA/RC/GA build
   is a human call.
3. **Take a new snapshot** of the updated VM with a fresh, unique name — Proxmox
   snapshots cannot be renamed — e.g. `Updated_to_25_10_5`. This is the new golden.
4. **Run the full test cycle** (`create → import → modify → destroy`, per
   `test-resources/README.md`) against it. The cycle starts from the golden, so it
   both confirms the snapshot boots and exercises the provider. A failure tied to a
   changed TrueNAS API shape is the signal that the provider needs updating for that
   release.

Repeat for each version. The versions are independent — refreshing the 25 golden
does not affect the 26 golden.

**Housekeeping:** snapshots form a branching tree — each refresh branches from the
golden you reverted to, so the 25 and 26 lines diverge from their common ancestor —
and Proxmox cannot rename a snapshot or delete one that still has children. Goldens
therefore accumulate and the chain grows over disk; periodically re-baseline from a
fresh `clean-install` lineage to keep it bounded.
