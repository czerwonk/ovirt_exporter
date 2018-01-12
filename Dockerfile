FROM golang

ENV API_INSECURE false
ENV WITH_SNAPSHOTS false
ENV WITH_NETWORK true

RUN apt-get install -y git && \
    go get github.com/czerwonk/ovirt_exporter

CMD ovirt_exporter -api.url=$API_URL -api.username=$API_USER -api.password=$API_PASS -api.insecure-cert=$API_INSECURE -with-snapshots=$WITH_SNAPSHOTS -with-network=$WITH_NETWORK
EXPOSE 9325
