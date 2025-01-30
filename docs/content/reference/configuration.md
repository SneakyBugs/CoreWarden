---
title: API server configuration
---

This page documents configuration options of the API server.

Configuration options can be set through command line flags, environment variables and configuration files in TOML, YAML, and JSON formats.

## `config`

Sets the configuration file. The API server spports reading YAML, TOML, and JSON configurations.
Defaults to `dns-api.*`.

Cannot be set from a configuration file, only from command line flag or environment variable.

Can be set through `DNSAPI_CONFIG` environment variable.


#### Example

Usage as command line flag:

```
api --config config-file.yaml
```

## `service-accounts`

Configures service accounts for authentication with the HTTP REST API.
Use id and secret configured for each service account as username and password for HTTP basic auth to authenticate to the API.

Each service account has an `id` and `secretHash`. `id` is used as the username, and `secretHash` is a Bcrypt hash used for validating the password for HTTP basic auth.

This configuration option **cannot be set through an environment variable.**

**Make sure to keep your secrets private.**
Do not use hashes from examples.
Avoid using online calculators to generate your hashes.

#### Example

Usage from YAML config:

```yaml
# Inside dns-api.yaml
service-accounts:
  # alice:secret
  - id: alice
    secretHash: $2a$10$LOXDp867yArcegMa/5TzxeHsvk/AiJBzWhK2tzHz4fNspvLQ7kPg6
  # bob:another-secret
  - id: bob
    secretHash: $2a$10$oxkrhLEncFfx.dCbdmpKk.oGX88sxA7T58k8VMNJ2Bgd4uk17UPCq
```

## `policy-file`

Sets the authorization policy file. Must be `csv` file [of the format specified in the Authorization refernece guide.]({{< relref "api-authorization" >}})
Defaults to `policy.csv`.

Can be set through `DNSAPI_POLICY_FILE` environment variable.

#### Example

Usage as command line flag:

```
api --policy-file policy.csv
```

Usage from YAML config:

```yaml
# Inside dns-api.yaml
policy-file: policy.csv
```

## `grpc-port`

Sets which port the gRPC server port is listening on.
Defaults to `6969`.

Can be set through `DNSAPI_GRPC_PORT` environment variable.

#### Example

Usage as command line flag:

```
api --grpc-port 8000
```

Usage from YAML config:

```yaml
# Inside dns-api.yaml
grpc-port: 8000
```

## `http-port`

Sets which port the HTTP server is listening on.
Defaults to `6970`.

Can be set through `DNSAPI_HTTP_PORT` environment variable.

#### Example

Usage as command line flag:

```
api --http-port 8001
```

Usage from YAML config:

```yaml
# Inside dns-api.yaml
http-port: 8001
```

## `postgres-host`

Sets the PostgreSQL database connection hostname.
**Required.**

Can be set through `DNSAPI_POSTGRES_HOST` environment variable.

#### Example

Usage as command line flag:

```
api --postgres-host 10.0.0.10
```

Usage from YAML config:

```yaml
# Inside dns-api.yaml
postgres-host: 10.0.0.10
```

## `postgres-port`

Sets the PostgreSQL database connection port.
Defaults to `5432`.

Can be set through `DNSAPI_POSTGRES_PORT` environment variable.

#### Example

Usage as command line flag:

```
api --postgres-port 5000
```

Usage from YAML config:

```yaml
# Inside dns-api.yaml
postgres-port: 5000
```

## `postgres-database`

Sets the PostgreSQL connection database name.
**Required.**

Can be set through `DNSAPI_POSTGRES_DATABASE` environment variable.

#### Example

Usage as command line flag:

```
api --postgres-database dnsapi
```

Usage from YAML config:

```yaml
# Inside dns-api.yaml
postgres-database: dnsapi
```

## `postgres-user`

Sets the PostgreSQL database connection user.
**Required.**

Can be set through `DNSAPI_POSTGRES_USER` environment variable.

#### Example

Usage as command line flag:

```
api --postgres-user dnsapi
```

Usage from YAML config:

```yaml
# Inside dns-api.yaml
postgres-user: dnsapi
```

## `postgres-password`

Sets the PostgreSQL database connection password.
**Required.**

Can be set through `DNSAPI_POSTGRES_PASSWORD` environment variable.

#### Example

Usage as command line flag:

```
api --postgres-password REDACTED
```

Usage from YAML config:

```yaml
# Inside dns-api.yaml
postgres-password: REDACTED
```
