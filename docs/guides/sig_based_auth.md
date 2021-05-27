---
subcategory: ""
page_title: "Signature-based Authentication"
description: |-
    An example of using signature-based authentication.
---

# ACI Authentication

Password-based authentication is very simple to work with, but it is not the most efficient form of authentication from ACI’s point-of-view as it requires a separate login-request and an open session to work. To avoid having your session time-out and requiring another login, you can use the more efficient Signature-based authentication.

Password-based authentication also may trigger anti-DoS measures in ACI v3.1+ that causes session throttling and results in HTTP 503 errors and login failures.

## Signature-based Authentication

Using signature-based authentication is more efficient and more reliable than password-based authentication.

### Generate certificate and private key

Signature-based authentication requires a (self-signed) X.509 certificate with private key, and a configuration step for your AAA user in ACI. To generate a working X.509 certificate and private key, use the following procedure:

```
$ openssl req -new -newkey rsa:1024 -days 36500 -nodes -x509 -keyout admin.key -out admin.crt -subj '/CN=Admin/O=Your Company/C=US'
```
### Configure your local user

Perform the following steps:

- Add the X.509 certificate to your ACI AAA local user at `ADMIN` » `AAA`
- Click AAA Authentication
- Check that in the `Authentication` field the `Realm` field displays `Local`
- Expand `Security Management` » `Local Users`
- Click the name of the user you want to add a certificate to, in the `User Certificates` area
- Click the `+` sign and in the `Create X509 Certificate` enter a certificate name in the `Name` field
    - If you use the basename of your private key here, you don’t need to enter `certificate_name` in Ansible
- Copy and paste your X.509 certificate in the `Data` field.

### Configure the provider

```terraform
provider "aci" {
  username = "admin"
  cert_name = "admin"
  private_key = "admin.key"
  url      = "https://10.1.1.1"
  insecure = true
}
```

The `cert_name` parameter must match the previously used certificate name when adding the certificate to the respective user.