FROM golang

ENV API_INSECURE false

RUN apt-get install -y git && \
    go get github.com/czerwonk/ovirt_exporter

CMD junos_exporter -api.url=$API_URL -api.username=$API_USER -api.password=$API_PASS -api.insecure-cert=$API_INSECURE
EXPOSE 9325
