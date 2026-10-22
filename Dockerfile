FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine3.24@sha256:d530cf60723abf1ffae0cd16bb546b379ec0550d2f3815e5b16ed69dbc00fba4 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG VERSION=dev

ADD . /go/ovirt_exporter/
WORKDIR /go/ovirt_exporter

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
  GOARM=${TARGETVARIANT#v} \
  go build -trimpath \
  -ldflags="-s -w -X main.version=${VERSION}" \
  -o /go/bin/ovirt_exporter .


FROM alpine:latest
ENV API_INSECURE false
ENV WITH_SNAPSHOTS false
ENV WITH_NETWORK true
ENV CMD_FLAGS ""
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /go/bin/ovirt_exporter .
CMD ./ovirt_exporter -api.url=$API_URL -api.username=$API_USER -api.password=$API_PASS -api.insecure-cert=$API_INSECURE -with-snapshots=$WITH_SNAPSHOTS -with-network=$WITH_NETWORK $CMD_FLAGS
EXPOSE 9325
