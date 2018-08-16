FROM golang as builder
RUN go get github.com/czerwonk/ovirt_exporter


FROM alpine:latest

ENV API_INSECURE false
ENV WITH_SNAPSHOTS false
ENV WITH_NETWORK true

RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/bin/ovirt_exporter .

CMD ovirt_exporter

EXPOSE 9325
