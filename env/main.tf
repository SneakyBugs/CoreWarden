terraform {
  required_providers {
    k3d = {
      source  = "sneakybugs/k3d"
      version = "1.0.1"
    }
  }
}

resource "k3d_cluster" "cluster" {
  name = "dns-dev"
  # See https://k3d.io/v5.4.6/usage/configfile/#config-options
  k3d_config = <<EOF
apiVersion: k3d.io/v1alpha5
kind: Simple

options:
  k3s:
    extraArgs:
      - arg: --disable=traefik
        nodeFilters:
          - server:*

# Expose ports 80 via 8080 and 443 via 8443.
ports:
  - port: 3080:80
    nodeFilters:
      - loadbalancer
  - port: 3443:443
    nodeFilters:
      - loadbalancer

registries:
  create:
    name: dev
    hostPort: "5000"
EOF
}

resource "local_sensitive_file" "kubeconfig" {
  content         = resource.k3d_cluster.cluster.kubeconfig
  filename        = "${path.module}/kubeconfig.yaml"
  file_permission = "0600"
}

provider "kubernetes" {
  host                   = resource.k3d_cluster.cluster.host
  client_certificate     = base64decode(resource.k3d_cluster.cluster.client_certificate)
  client_key             = base64decode(resource.k3d_cluster.cluster.client_key)
  cluster_ca_certificate = base64decode(resource.k3d_cluster.cluster.cluster_ca_certificate)
}

provider "helm" {
  kubernetes {
    host                   = resource.k3d_cluster.cluster.host
    client_certificate     = base64decode(resource.k3d_cluster.cluster.client_certificate)
    client_key             = base64decode(resource.k3d_cluster.cluster.client_key)
    cluster_ca_certificate = base64decode(resource.k3d_cluster.cluster.cluster_ca_certificate)
  }
}

resource "kubernetes_secret" "postgres_credentials" {
  metadata {
    name = "postgres-credentials"
  }

  data = {
    "postgres-password"    = "development"
    "password"             = "development"
    "replication-password" = "development"
    "username"             = "development"
  }
}

resource "helm_release" "traefik" {
  name             = "traefik"
  repository       = "oci://ghcr.io/traefik/helm"
  chart            = "traefik"
  namespace        = "traefik-system"
  create_namespace = true
  wait             = true

  set {
    name  = "providers.kubernetesGateway.enabled"
    value = true
  }
}

resource "helm_release" "cloudnative_pg" {
  name             = "cloudnative-pg"
  repository       = "https://cloudnative-pg.github.io/charts"
  chart            = "cloudnative-pg"
  namespace        = "cnpg-system"
  create_namespace = true
  wait             = true
}

resource "helm_release" "pg" {
  depends_on = [helm_release.cloudnative_pg]
  name       = "corewarden"
  repository = "https://cloudnative-pg.github.io/charts"
  chart      = "cluster"
  wait       = true
  values = [
    <<EOF
cluster:
  initdb:
    database: corewarden
  instances: 1
  resources:
    requests:
      cpu: "250m"
      memory: "512Mi"
    limits:
      cpu: "250m"
      memory: "512Mi"
EOF
  ]
}
