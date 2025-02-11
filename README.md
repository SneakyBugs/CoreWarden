<p align="center">
  <a href="https://sneakybugs.github.io/CoreWarden/">
    <img width="65%" src="https://sneakybugs.github.io/CoreWarden/images/widetitle.svg">
  </a>
</p>

<p align="center">
  CoreWarden is the Kubernetes-native ad-blocking DNS server for your homelab.
  <a href="https://sneakybugs.github.io/CoreWarden/">
    Check out the documentation.
  </a>
</p>

<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#getting-started">Getting Started</a> •
  <a href="#contribution">Contribution</a>
</p>

## Key features

- Headless operation, designed to be managed via configs and API.
- Supports high availability and rolling upgrades.
- Fine grained authorization for record editing.
- Built-in DOT support for encrypted DNS lookups.
- Built-in Prometheus metrics exporting and structured logging for easy use with observability tools.
- Official External DNS provider.

## What it is not

- It is not an authoritative server. It is not designed for being the nameserver for your domains.
- It is not a recursive resolver.

## Why not Adguard, PiHole or Technitium?

I've been using PiHole and Adguard for the past few years.
However, I found them less than ideal for running on Kubernetes.

- They don't provide Helm charts or Kubernetes manifests.
- They don't export Prometheus metrics.
- They aren't designed for high availability and rolling deployments.

[Read more about the reasoning in the documentation.](https://sneakybugs.github.io/CoreWarden/explanations/requirements/#why)

## Getting started

[Read the getting started guide for installation on Kubernetes with Helm.](https://sneakybugs.github.io/CoreWarden/tutorials/installation-kubernetes/)

## Contribution

[See the development instructions in the documentation.](https://sneakybugs.github.io/CoreWarden/explanations/development/)
If you want to learn more about how the project works,
[check out the architecture page.](https://sneakybugs.github.io/CoreWarden/explanations/architecture/)
