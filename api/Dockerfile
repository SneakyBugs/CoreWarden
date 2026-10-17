# Needs to be built from the root directory.
FROM golang:1.26.4-bookworm AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY api api
RUN cd api && go build .

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build /build/api /
ENTRYPOINT ["/api"]
