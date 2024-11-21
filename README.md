# ovirt_exporter
[![Go Report Card](https://goreportcard.com/badge/github.com/czerwonk/ovirt_exporter)](https://goreportcard.com/report/github.com/czerwonk/ovirt_exporter)

Exporter for oVirt engine metrics to use with https://prometheus.io/

This project is in maintenance mode now. No new features will be implemented by myself, merge requests are welcome.

## Install
```
go get -u github.com/czerwonk/ovirt_exporter
```

## Supported ressources
* hosts
* vms
* storagedomains
* snapshots (optional)

## Third Party Components
This software uses components of the following projects
* Prometheus Go client library (https://github.com/prometheus/client_golang)

## License
(c) Daniel Czerwonk, 2017. Licensed under [MIT](LICENSE) license.

## Prometheus
see https://prometheus.io/

## oVirt
see https://www.ovirt.org/
