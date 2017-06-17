# ovirt_exporter
[![Build Status](https://travis-ci.org/czerwonk/ovirt_exporter.svg)][travis]
[![Go Report Card](https://goreportcard.com/badge/github.com/czerwonk/ovirt_exporter)][goreportcard]

Exporter for oVirt engine metrics to use with https://prometheus.io/

## Remarks
this is an early version

## Install
```
go get -u github.com/czerwonk/ovirt_exporter
```

## Supported ressources
* hosts
* vms

## Third Party Components
This software uses components of the following projects
* Prometheus Go client library (https://github.com/prometheus/client_golang)

## License
(c) Daniel Czerwonk, 2017. Licensed under [MIT](LICENSE) license.

## Prometheus
see https://prometheus.io/

## oVirt
see https://www.ovirt.org/

[travis]: https://travis-ci.org/czerwonk/ovirt_exporter
[goreportcard]: https://goreportcard.com/report/github.com/czerwonk/ovirt_exporter
