---
page_title: "trueform_cronjob Resource - Trueform"
subcategory: "System"
description: |-
  Manages a cron job on TrueNAS.
---

# trueform_cronjob (Resource)

Manages a cron job on TrueNAS Scale for scheduled task execution.

## Example Usage

### Daily Backup Script

```hcl
resource "trueform_cronjob" "daily_backup" {
  user        = "root"
  command     = "/usr/local/bin/backup.sh"
  description = "Daily backup at 2 AM"
  enabled     = true

  schedule = {
    minute = "0"
    hour   = "2"
    dom    = "*"
    month  = "*"
    dow    = "*"
  }
}
```

### Weekly Maintenance

```hcl
resource "trueform_cronjob" "weekly_scrub" {
  user        = "root"
  command     = "zpool scrub tank"
  description = "Weekly pool scrub on Sunday"
  enabled     = true

  schedule = {
    minute = "0"
    hour   = "0"
    dom    = "*"
    month  = "*"
    dow    = "0"  # Sunday
  }

  stdout = true
  stderr = true
}
```

### Hourly Task

```hcl
resource "trueform_cronjob" "hourly_sync" {
  user        = "root"
  command     = "/usr/local/bin/sync.sh"
  description = "Hourly sync"
  enabled     = true

  schedule = {
    minute = "0"
    hour   = "*"
    dom    = "*"
    month  = "*"
    dow    = "*"
  }
}
```

## Schema

### Required

- `command` (String) Command to execute.
- `schedule` (Object) Cron schedule configuration.
  - `minute` (String) Minute (0-59 or `*`).
  - `hour` (String) Hour (0-23 or `*`).
  - `dom` (String) Day of month (1-31 or `*`).
  - `month` (String) Month (1-12 or `*`).
  - `dow` (String) Day of week (0-6, where 0=Sunday, or `*`).
- `user` (String) User to run the command as.

### Optional

- `description` (String) Job description.
- `enabled` (Boolean) Enable the cron job. Defaults to `true`.
- `stderr` (Boolean) Redirect stderr to syslog. Defaults to `true`.
- `stdout` (Boolean) Redirect stdout to syslog. Defaults to `true`.

### Read-Only

- `id` (Number) Cron job identifier.

## Import

Cron jobs can be imported using the job ID:

```shell
terraform import trueform_cronjob.daily_backup 1
```
