---
title: How-to use ExternalDNS for record management
---

In this guide we will go through setting up the ExternalDNS integration.
After following this guide, you will be able to dynamically set DNS records for Kubernetes Services, Ingresses, and other resources with ExternalDNS.

## Requirements

1. DNS server running and set as your computer's primary DNS.
1. Service account for the DNS API server with read and write permissions for the zone you want to manage records in.
1. Access to a Kubernetes cluster. Does not have to be the same cluster running the DNS server. `kind` or `k3d` may be used for evaluation purposes.

## Setting up ExternalDNS

Start by editing the following `ConfigMap` and `Secret` to fit your DNS API server endpoint, zones, and service account.

```yaml
# Inside config.yml
---
apiVersion: v1
kind: Namespace
metadata:
  name: external-dns
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: webhook
  namespace: external-dns
data:
  apiEndpoint: http://api.default:6970/v1
  zones: example.com.
---
apiVersion: v1
kind: Secret
metadata:
  name: webhook-credentials
  namespace: external-dns
stringData:
  id: alice
  secret: debug
```

Apply the manifests:

```
kubectl apply -f config.yml
```

Create a values file for the ExternalDNS Helm chart:

```yaml
# Inside values.yaml
provider:
  name: webhook
  webhook:
    image:
      repository: oci://ghcr.io/sneakybugs/corewarden-externaldns-provider
      tag: 3.0.0
    env:
      - name: WEBHOOK_API_ENDPOINT
        valueFrom:
          configMapKeyRef:
            name: webhook
            key: apiEndpoint
      - name: WEBHOOK_API_ID
        valueFrom:
          secretKeyRef:
            name: webhook-credentials
            key: id
      - name: WEBHOOK_API_SECRET
        valueFrom:
          secretKeyRef:
            name: webhook-credentials
            key: secret
      - name: WEBHOOK_ZONES
        valueFrom:
          configMapKeyRef:
            name: webhook
            key: zones
    livenessProbe:
      httpGet:
        path: /-/liveness
    readinessProbe:
      httpGet:
        path: /-/readiness
```

Add the ExternalDNS Helm repository:

```
helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
```

Install the ExternalDNS Helm chart with the values file we created:

```
helm upgrade --install external-dns external-dns/external-dns -f values.yaml
```

ExternalDNS is now ready to use. Make sure the ExternalDNS pod is up and running:

```
kubectl get pods -n external-dns
```

## Checking it out

Create the following manifest deploying an Nginx Pod and Service.

```yaml
# Inside nginx.yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: debug
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: debug
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
  namespace: debug
  annotations:
    external-dns.alpha.kubernetes.io/hostname: my-app.example.com
spec:
  selector:
    app: nginx
  type: LoadBalancer
  ports:
    - protocol: TCP
      port: 8000
      targetPort: 80
```

Apply the manifests:

```
kubectl apply -f nginx.yaml
```

Verify the DNS server responds with the updated record:

```
dig my-app.example.com
```

## Troubleshooting

First make sure the DNS server responds correctly to other domain names:

```
dig google.com
```

If it does not respond correctly your issue is not related to ExternalDNS but to the DNS server.

If the DNS server does respond correctly, the issue is likely related to communication between the ExternalDNS webhook and the DNS API server.
Check out the webhook logs and search for error messages.

Find the name of the ExternalDNS Pod by listing pods in the `external-dns` namespace:

```
kubectl get pods -n external-dns
```

View the logs of the `webhook` container in the ExternalDNS Pod:

```
kubectl logs -n external-dns external-dns-xxxxxxxxxx-xxxxx webhook
```

If you do not see what's wrong, you will likely find it in the DNS API server logs:

```
kubectl logs dns-api-xxxxxxxxx-xxxxx
```

If you do not see any attempts to set records, the issue is likely with ExternalDNS itself.
