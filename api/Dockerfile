FROM golang:1.21.3-bookworm AS build

WORKDIR /api

COPY . .
RUN go build .

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build /api/api /
ENTRYPOINT ["/api"]
