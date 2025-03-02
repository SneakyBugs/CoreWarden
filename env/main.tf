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

resource "helm_release" "database" {
  name       = "postgres"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "postgresql"
  set {
    name  = "auth.existingSecret"
    value = "postgres-credentials"
  }
  set {
    name  = "auth.username"
    value = "development"
  }
  set {
    name  = "auth.database"
    value = "development"
  }
}
