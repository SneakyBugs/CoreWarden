# Needs to be built from the root directory.
FROM golang:1.26.4-bookworm AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY api api
COPY client client
COPY coredns/plugin/injector/zones coredns/plugin/injector/zones
COPY external-dns external-dns
RUN cd external-dns && go build .

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build /build/external-dns/external-dns /webhook
ENTRYPOINT ["/webhook"]
