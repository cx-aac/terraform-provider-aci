---
page_title: "aci_rest Data Source - terraform-provider-aci"
subcategory: ""
description: |-
  This data source can read one ACI object and its children.
---

# Data Source `aci_rest`

This data source can read one ACI object and its children.

## Example Usage

```terraform
data "aci_rest" "fvTenant" {
  dn = "uni/tn-EXAMPLE_TENANT"
}
```

## Schema

### Required

- **dn** (String) Distinguished name of object to be retrieved, e.g. uni/tn-EXAMPLE_TENANT.

### Read-only

- **child** (Set of Object) Set of children of object being retrieved. (see [below for nested schema](#nestedatt--child))
- **class_name** (String) Class name of object being retrieved.
- **content** (Map of String) Map of key-value pairs which represents the attributes of object being retrieved.
- **id** (String) The distinguished name of the object.

<a id="nestedatt--child"></a>
### Nested Schema for `child`

Read-only:

- **class_name** (String)
- **content** (Map of String)


