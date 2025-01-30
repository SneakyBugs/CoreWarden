---
title: Requirements
---

The cloud-native DNS server for your homelab.

## Why

I need a DNS server to perform both as an ad blocker, and a resolver for a split horizon DNS setup.

I've been running Pihole and Adguard in my homelab for years and encountered numerous headaches:

- Adguard, Pihole, and Technitium do not provide official Helm charts or Kubernetes manifests.
  Making the alternatives hard to install and manage on Kubernetes.
- Adguard and Pihole do not support rewriting of all record types.
  Causing automations such as ExternalDNS providers to either use hacky solutions or support very few record types.
- Adguard and Pihole are not designed to run in headless, stateless containers.
  Making them hard to manage using automated tooling.
- [Adguard has been ignoring requests for metrics endpoint for years.](https://github.com/AdguardTeam/AdGuardHome/issues/516)
  Recently the most popular third party Prometheus exporter vanished from GitHub.
  Making Adguard very hard to collect metrics from.
- Adguard config requires write access, meaning mounted configmaps cannot be used.
  [Adguard has been ignoring related issues for years.](https://github.com/AdguardTeam/AdGuardHome/issues/1964)
- Pihole cannot rewrite wildcard records without editing the dnsmasq config.
  Making it hard to configure wildcard records automatically.
- Pihole does not support DOH and DNSSEC for upstream DNS.
  Requiring users to manage a DOH proxy sidecar if they want to encrypt requests to their upstream.
- Adguard, Pihole, and Technitium are designed to store request logs and metrics internally.
  Making it hard to integrate with centralized observability systems.
- Adblock style blocklists in Adguard allow DNS injection through blocklists.
  Making blocklists a potential attack vector for DNS injection.

## What it is not

- It is not an authoritative server. It is not for being the root nameserver for your domains.
- It is not a recursive resolver. It will connect to an upstream DNS server using an encrypted protocol.

## Requirements

1. Must support ad blocking with Pihole or Adguard formatted blocklists.
1. Must support caching of DNS queries.
1. Must support DOT or DOH upstream.
1. Must support DNS record rewrites.
1. Must support DNS record rewrite management through API.
1. Must support API authentication with service accounts consisting of an ID and secret pair.
1. Must support fine-grained authorization for API actions.
1. Must have readiness and liveness endpoints for all services.
1. Must export structured DNS query log.
1. Must export Prometheus or OpenTelemetry metrics.
1. Must export OpenTelemetry traces.
1. Must supply a Helm chart.
1. Must supply a Go API client.
1. Must supply an ExternalDNS webhook provider.
1. Must support API authentication with Kubernetes service accounts.

### First iteration

Released Dec 2nd, 2023.

- Ad blocking (1)
- Caching (2)
- DOT upstream (3)
- Rewrites (4)
- Probes (8)
- Structured log (9)
- Metrics (10)
- Helm chart (12)

### Second iteration

Released May 16th, 2024.

- Rewrites API (5)
- Service account auth (6)
- Authorization (7)

### Third iteration

Released July 1st, 2024.

- API client (13)
- ExternalDNS provider (14)

## Progress

| Requirement | MH/NTH | Risk | Status |
|-|-|-|-|
| Ad blocking (1) | MH | High | Done |
| Caching (2) | MH | Low | Done |
| DOT upstream (3) | MH | Low | Done |
| Rewrites (4) | MH | Low | Done |
| Rewrites API (5) | MH | High | Done |
| Service account auth (6) | MH | Low | Done |
| Authorization (7) | MH | High | Done |
| Probes (8) | MH | Low | Done |
| Structured log (9) | Low | MH | Done |
| Metrics (10) | MH | Low | Done |
| Traces (11) | MH | High | Not done |
| Helm chart (12) | MH | Low | Done |
| API client (13) | MH | Low | Done |
| ExternalDNS provider (14) | NTH | High | Done |
| Kubernetes auth (15) | NTH | High | Not done |
