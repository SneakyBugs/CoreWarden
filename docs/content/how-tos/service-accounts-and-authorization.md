---
title: How-to create a service account with authorization
---

In this guide we will create a service account and authorize it for editing
records in a DNS zone.
After following this guide, you will have credentials you can use with the record
management API.

## Requirements

1. DNS server running and set as your computer's primary DNS.
2. DNS server is installed using Helm.

## Creating the service account

In this example we'll create a user with `id` of `example-user` and
`secret` of `example-password`.

```yaml
# Inside dns-api-values.yaml
config:
  serviceAccounts:
    - id: example-user
      secretHash: $2a$10$CeuWZl38oQi0iX6yMXqgf.pNYD4Vod.FtyxCSWSToSsoNx2z/sPuO
```

## Authorizing the service account for editing records

We want to authorize the `example-user` service account for editing records on
`example.com.` zone.
Note the dot in the end of the zone name is required.

```yaml
# Inside dns-api-values.yaml
config:
  # ...
  policies: |-
    p, example-user, records, example.com., read
    p, example-user, records, example.com., edit
```

## Applying the config

Apply the API server Helm chart with the updated values to update the API server
configuration:

```yaml
helm upgrade --install dnsapi-api TODO --values dns-api-values.yaml
```

## Verifying the service account works

Replace `<dns-api-server>` with the url of your DNS API server.

```
curl -u example-user:example-password '<dns-api-server>/v1/records?zone=example.com.'
```

You should see a list of records in the `example.com.` zone.
