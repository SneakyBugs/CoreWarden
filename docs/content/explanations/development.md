---
title: Development instructions
---

This guide covers all commands needed for development.

## Using the development environment

This project uses a K3d cluster and Tilt for development.
The `env` directory contains a Terraform module for bringing up the K3d cluster
and configuring required resources on it.

Go into the `env` directory:

```
cd env
```

Initialize the OpenTofu module:

```
tofu init
```

The Gateway resources need the cluster and Gateway API CRDs to exist at plan
time. On a fresh environment, first create the cluster and install Envoy Gateway:

```
tofu apply -target=helm_release.envoy_gateway
```

Apply the full module (a single step for all subsequent applies):

```
tofu apply
```

Set the `KUBECONFIG` environment variable to point to the kubeconfig file to
configure `kubectl` and Tilt to use it.

```
export KUBECONFIG=$(pwd)/kubeconfig.yaml
```

Go back to the project root directory and bring Tilt up:

```
cd ..
tilt up
```

Open the Tilt dashboard in your browser.
You now have a cluster running and automatic rebuilds for all parts of the
project.

## Generating Sqlc code

The API server in the `api` directory uses Sqlc for performing SQL queries.
Sqlc generates type safe code from SQL.

Install Sqlc:

```
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.24.0
```

Go to the API server directory:

```
cd api
```

Generate Sqlc code:

```
sqlc generate
```

## Generating gRPC code

Both the API server and Injector CoreDNS plugin use generated code for gRPC client
and server.

Install dependenices for generating gRPC code:

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

Go to the `proto` directory containing Protobuf definitions:

```
cd proto
```

Generate gRPC server and client code for both the API server and Injector CoreDNS
plugin:

```
protoc --go_out=../api/resolver --go_opt=paths=source_relative --go-grpc_out=../api/resolver --go-grpc_opt=paths=source_relative resolver.proto
protoc --go_out=../coredns/plugin/injector/resolver --go_opt=paths=source_relative --go-grpc_out=../coredns/plugin/injector/resolver --go-grpc_opt=paths=source_relative resolver.proto
```

## Running tests

A small subset of API server tests require access to a Postgres database.
Use the following commands in a separate shell to port-forward the development
Postgres database to the host:

```
export KUBECONFIG=$(pwd)/env/kubeconfig.yaml
kubectl port-forward postgres-postgresql-0 5432
```

Run tests:

```
go test ./...
```

## Linting and formatting

Lint the source files:

```
golangci-lint run ./...
```

Format the source files:

```
go fmt ./...
```
