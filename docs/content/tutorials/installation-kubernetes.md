---
title: Getting started on Kubernetes using Helm
---

In this guide we will install CoreWarden on a local Kubernetes cluster.

You may follow this guide using a production Kubernetes cluster.
Following the notes for production deployment, this guide can be used for
configuring a robust production setup.

## Setting up a cluster

To follow this tutorial you will need a Kubernetes cluster.
In this section we will set up a local cluster with `k3d` which will work fine for evaluation or development purposes.
[Skip to the next section if you already have a cluster you want to install on.](#installation)

[Begin by installing `k3d`.](https://k3d.io/#installation)
Then create the cluster and configure `kubectl` to use the cluster:


```
sudo k3d cluster create dns-tutorial
sudo k3d kubeconfig merge dns-tutorial --output ~/.kube/config
sudo chown <you> ~/.kube/config
```

Now let's check the cluster works.
List all pods:

```
kubectl get pods -A
```

Ensure the output looks similar to this:

```
NAMESPACE     NAME                                      READY   STATUS      RESTARTS   AGE
kube-system   coredns-6799fbcd5-wd2qx                   1/1     Running     0          2m56s
kube-system   helm-install-traefik-crd-94mwm            0/1     Completed   0          2m56s
kube-system   svclb-traefik-57c533c5-5wz6w              2/2     Running     0          2m46s
kube-system   helm-install-traefik-sgvp4                0/1     Completed   1          2m56s
kube-system   traefik-f4564c4f4-ldbgv                   1/1     Running     0          2m46s
kube-system   local-path-provisioner-6c86858495-4zzkh   1/1     Running     0          2m56s
kube-system   metrics-server-54fd9b65b-kktdb            1/1     Running     0          2m56s
```

You now have a Kubernetes cluster up and running.

## Installation

### Database setup

In this step we will install the Bitnami Postgres chart to quickly set up a Postgres database.

Create a Helm values file named `postgres-values.yaml` with authentication credentials:

```yaml
# Inside postgres-values.yaml
auth:
  username: api
  database: api
  password: secret_value
```

Install the Postgres Helm chart providing the values file created in the previous step:

```
helm install corewarden-db oci://registry-1.docker.io/bitnamicharts/postgresql --values postgres-values.yaml
```

#### Notes for production deployment

- Use a managed database solution. CloudNativePG is recommended for self managed deployments.
- Avoid commiting secrets to version control.

### API server installation

Create a values file named `api-values.yaml` with the following content:

```yaml
# Inside api-values.yaml
config:
  policies: |-
    p, tutorial, records, example.com., edit
    p, tutorial, records, example.com., read
  serviceAccounts:
    # tutorial:example
    - id: tutorial
      secretHash: $2a$10$ksbGVKQ6MBjH9vuKWHEloOwOHdFBEX6abYRtnOmat.camf2ogIrmq
  postgres:
    host: corewarden-db-postgresql
    database: api
    existingSecret:
      name: corewarden-db-credentials
ingress:
  host: dns.example.com
```

Install the `api` chart:

```
helm install corewarden-api oci://ghcr.io/sneakybugs/corewarden-api --values api-values.yaml
```

#### Notes for production deployment

It is recommended to use Cert Manager to issue TLS certificates for the API server:

```yaml
# Inside api-values.yaml
ingress:
  annotations:
    cert-manager.io/cluster-issuer: example-issuer
```

### DNS server installation

Create a values file named `coredns-values.yaml` with the following content:

```yaml
# Inside coredns-values.yaml
config:
  injectorTarget: corewarden-api:6969
```

Install the `coredns` chart:

```
helm install corewarden-coredns oci://ghcr.io/sneakybugs/corewarden-coredns --values coredns-values.yaml
```

#### Notes for production deployment

On production clusters you may need additional annotations to allow the DNS service to use a single IP for serving UDP and TCP on port 53.
For example when using MetalLB:

```yaml
# Inside coredns-values.yaml
service:
  annotations:
    metallb.universe.tf/allow-shared-ip: dns
```

## Summary and next steps

We have successfully deployed an example deployment of CoreWarden.
Make sure to follow the production deployment notes for a robust production setup.

The next step would be [configuring External DNS on your Kubernetes cluster
by following this guide.]({{< relref "../how-tos/using-external-dns" >}})
So that DNS records will automatically be created for services running on
Kubernetes.
