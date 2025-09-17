# Needs to be built from the root directory.
FROM golang:1.25.1-bookworm AS build

WORKDIR /build

COPY go.mod go.sum ./
COPY api api
RUN cd api && go build .

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build /build/api /
ENTRYPOINT ["/api"]
