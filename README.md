# Podnoise exporter

Prometheus exporter for pod noise metrics measured with bash scripts.

[![Go Report Card](https://goreportcard.com/badge/github.com/stedc1976/bash-exporter)](https://goreportcard.com/badge/github.com/stedc1976/bash-exporter)

## Installation

```console
$ docker build --rm -t diclem27/podnoise-exporter:1.0.0 .
```

## Docker quick start

```console
$ docker run -d -p 9300:9300 --name my_podnoise-exporter diclem27/podnoise-exporter:1.0.0
```

```console
$ curl -s 127.0.0.1:9300/metrics | grep logrowcount
logrowcount{container_log="3ds-card-4038631ff5792525eee47ff9c0cb945da44cd05f9d48d943fb010d747655b970.log",namespace="psd2",pod_name="3ds-card-12-7spjp"} 222
logrowcount{container_log="redis-60faea9b6ec32975946402567018784ab4783e6370c10c7934916f62a80c5157.log",namespace="crmu-dev",pod_name="redis-3-t4qlh"} 8
```

## Usage

```console
Usage of ./podnoise-exporter:
  -debug
    	if true, print log messages to stdout (default false)
  -interval int
    	interval for metrics collection in seconds (default 10)
  -pathname string
    	pathname bash script (default "./scripts/job.sh")
  -web.listen-address string
    	address on which to expose metrics (default ":9300")
```

bash script should return valid json.

## External doc

https://godoc.org/github.com/prometheus/client_golang/prometheus

## TODO
- [] Helm Chart