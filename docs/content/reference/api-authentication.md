---
title: API authentication
---

The REST API supports authentication through service accounts and HTTP basic auth.

Service accounts are defined in the API server configuration.
Then the service account ID and secret are used as username and password in HTTP basic auth to authenticate with the API.

## Defining service accounts with config file

Service accounts are defined
[through the `service-accounts` config option.]({{< relref "configuration#service-accounts" >}})

For example in a YAML config:

```yaml
# Inside dns-api.yaml
service-accounts:
  # alice:secret
  - id: alice
    secretHash: $2a$10$LOXDp867yArcegMa/5TzxeHsvk/AiJBzWhK2tzHz4fNspvLQ7kPg6
```

## Defining service accounts with Helm chart values

When deploying with Helm you will want to define the service accounts through Helm values.

Service accounts can be defined in the following format in API Helm chart values:

```yaml
config:
  serviceAccounts:
    # alice:secret
    - id: alice
      secretHash: $2a$10$LOXDp867yArcegMa/5TzxeHsvk/AiJBzWhK2tzHz4fNspvLQ7kPg6
```

## Authenticating with the API

Use HTTP basic auth with the service account ID and secret as username and password respectively.

For example listing records as `alice` defined in the above examples:

```bash
curl -u alice:secret http://dns.example.com/v1/records\?zone\=example.com.
```
