---
title: Architecture
---

The repository contains 4 distinct parts in subdirectories:

- `api` the record editing API server
- `client` Go client for the record editing API
- `coredns` customized CoreDNS build
- `external-dns` webhook for usage with External DNS operator

![](../architecture.svg)

## API server and client

The `api/openapi.yaml` file contains an OpenAPI specification for the record
editing API used by the API server and client.
Both the server and client are tested against this specification.

The API server manages record overrides.
On each DNS query, CoreDNS communicates with the API server with gRPC to query for
record overrides. CoreDNS then returns the overridden record if there is one.

## CoreDNS

CoreDNS is used as the DNS server being queried by users.
It is separate from the API server to allow for operation without the API server,
and for operating multiple DNS servers from the same API server (for example one
server with ad-blocking and one without).

The `coredns/plugin/filterlist` directory contains the CoreDNS plugin implementing
ad-blocking logic.

The `coredns/plugin/injector` directory contains the CoreDNS plugin implementing
lookups in the API server over gRPC.

## External DNS webhook

The `external-dns` directory contains a webhook for External DNS operator.
This webhook uses the API client to automatically update DNS records for services
running on Kubernetes.

Using External DNS, multiple clusters can automatically update DNS records in the
API server seamlessly.
