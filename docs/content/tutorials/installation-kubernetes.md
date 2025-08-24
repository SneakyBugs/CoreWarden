---
title: Getting started on Kubernetes using Helm
---

In this guide we will install CoreWarden on a local Kubernetes cluster.

You may follow this guide using a production Kubernetes cluster.
Following the notes for production deployment, this guide can be used for
configuring a robust production setup.

## Requirements for following this guide

You will need a Linux system with the following installed:

- dig
- Helm
- kubectl
- k3d (if you don't already have a Kubernetes cluster)

## Setting up a cluster

To follow this tutorial you will need a Kubernetes cluster.
In this section we will set up a local cluster with `k3d` which will work fine for evaluation or development purposes.
[Skip to the next section if you already have a cluster you want to install on.](#installation)

Create a cluster and configure `kubectl` to use the cluster:


```
sudo k3d cluster create dns-tutorial
sudo k3d kubeconfig merge dns-tutorial --output ~/.kube/config
sudo chown $(whoami) ~/.kube/config
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

## Installing CoreWarden

### Database installation

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
helm upgrade --install corewarden-db oci://registry-1.docker.io/bitnamicharts/postgresql --values postgres-values.yaml
```

#### Notes for production deployment

- Use a managed database solution. CloudNativePG is recommended for self managed deployments.
- Avoid commiting secrets to version control.

### API server installation

Create a values file named `api-values.yaml` with the following content:

```yaml
# Inside api-values.yaml
fullnameOverride: corewarden-api
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
    user: api
    password: secret_value
ingress:
  enabled: true
  host: dns.example.com
```

Install the `api` chart:

```
helm upgrade --install corewarden-api oci://ghcr.io/sneakybugs/corewarden-api-chart --values api-values.yaml
```

#### Notes for production deployment

It is recommended to use Cert Manager to issue TLS certificates for the API server:

```yaml
# Inside api-values.yaml
ingress:
  annotations:
    cert-manager.io/cluster-issuer: example-issuer
```

### CoreDNS server installation

Create a values file named `coredns-values.yaml` with the following content:

```yaml
# Inside coredns-values.yaml
fullnameOverride: corewarden-coredns
config:
  injectorTarget: corewarden-api:6969
```

Install the `coredns` chart:

```
helm upgrade --install corewarden-coredns oci://ghcr.io/sneakybugs/corewarden-coredns-chart --values coredns-values.yaml
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

## Verifying the DNS is working

Get the Service used for DNS and figure out it's external IP:

```
kubectl get service corewarden-coredns-tcp
```

In our case the external IP is `172.19.0.2`. Let's perform a DNS query to check it
works:

```
dig @172.19.0.2 google.com
```

The output should look like:

```
; <<>> DiG 9.18.33-1~deb12u2-Debian <<>> @172.19.0.2 google.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 14892
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
; COOKIE: f510bec38f682e60 (echoed)
;; QUESTION SECTION:
;google.com.                    IN      A

;; ANSWER SECTION:
google.com.             195     IN      A       142.250.75.142

;; Query time: 0 msec
;; SERVER: 172.19.0.2#53(172.19.0.2) (UDP)
;; WHEN: Tue Feb 11 18:47:16 UTC 2025
;; MSG SIZE  rcvd: 77
```

We have verified our DNS server works.

## Summary and next steps

We have successfully deployed an example deployment of CoreWarden.
Make sure to follow the production deployment notes for a robust production setup.

The next step would be [configuring External DNS on your Kubernetes cluster
by following this guide.]({{< relref "../how-tos/using-external-dns" >}})
So DNS records will automatically be created for services running on Kubernetes.
