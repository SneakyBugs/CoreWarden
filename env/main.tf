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
apiVersion: k3d.io/v1alpha4
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

# K3d uses a hard coded DNS server.
# This resource causes it to use our private DNS.
# See https://github.com/k3s-io/k3s/pull/4397
# Private DNS is required for using private Gitlab as OIDC provider.
resource "kubernetes_config_map" "coredns" {
  metadata {
    name      = "coredns-custom"
    namespace = "kube-system"
  }
  data = {
    "private.server" = <<EOF
houseofkummer.com:53 {
  forward . 192.168.0.180:53
}
EOF
  }
}
