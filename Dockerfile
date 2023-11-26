FROM golang:1.21.4-bookworm AS build

WORKDIR /build

RUN wget -O coredns.tar.gz https://github.com/coredns/coredns/archive/refs/tags/v1.11.1.tar.gz \
	&& tar -xzf coredns.tar.gz \
	&& cp -r coredns-* coredns \
	&& rm -rf coredns-*

COPY plugin.cfg /build/coredns/plugin.cfg
COPY filterlist /build/coredns/plugin/filterlist
COPY slog /build/coredns/plugin/slog

RUN cd coredns && go mod tidy && go generate && go build

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build /build/coredns/coredns /
ENTRYPOINT ["/coredns"]
