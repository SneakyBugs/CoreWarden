FROM golang:1.22.3-bookworm AS build

WORKDIR /coredns

COPY Makefile ./
RUN make build/coredns

COPY plugin.cfg ./
COPY plugin plugin
RUN make coredns

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build /coredns/coredns /
ENTRYPOINT ["/coredns"]
