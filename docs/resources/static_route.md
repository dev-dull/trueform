---
page_title: "trueform_static_route Resource - Trueform"
subcategory: "Network"
description: |-
  Manages a static network route on TrueNAS.
---

# trueform_static_route (Resource)

Manages a static network route on TrueNAS Scale.

## Example Usage

```hcl
resource "trueform_static_route" "internal" {
  destination = "10.0.0.0/8"
  gateway     = "192.168.1.1"
  description = "Route to internal network"
}
```

## Schema

### Required

- `destination` (String) Destination network in CIDR notation (e.g., `10.0.0.0/8`).
- `gateway` (String) Gateway IP address.

### Optional

- `description` (String) Route description.

### Read-Only

- `id` (Number) Route identifier.

## Import

Static routes can be imported using the route ID:

```shell
terraform import trueform_static_route.internal 1
```
