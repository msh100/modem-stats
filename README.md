# Modem Statistics Parser

A utility to read and parse channel diagnostics information from DOCSIS and
VDSL modems.
This package is intended to be used within a Telegraf instance to be fed into
InfluxDB.
A Prometheus endpoint can also be exposed for collection.

This package has been written in Go in an attempt to allow it to run on low end
hardware (such as a Raspberry Pi Zero) with no issues.

![Grafana dashboard screenshot](https://user-images.githubusercontent.com/4477262/114266746-cd910980-99ef-11eb-8a5e-f4f719897719.JPG)


 * [Usage](#Usage)
   * [Docker Image](#Docker-Image)
   * [Telegraf](#Telegraf)
   * [Prometheus](#Prometheus)
 * [Binaries](#Binaries)
   * [Download](#Downloading)
   * [Build](#Building)
 * [Configuration](#Configuration)
   * [Example Usage](#Example-Usage)
 * [Grafana](#Grafana)


## Usage

In its simplest form, you can run this repository directly.
The only dependency being Go.

A compiled binary of this repository will require no dependencies.


### Docker Image

A Docker image exists for this repository with Telegraf configured to write to
InfluxDB at `msh100/modem-stats`.
This image currently only supports X86.

The following environment variables must be set:

Name             | Description                                 | Example
-----------------|---------------------------------------------|---------------------
`INFLUX_URL`     | The HTTP API URI for your InfluxDB server.  | `http://influxdb:8086`
`INFLUX_DB`      | The InfluxDB database to use                | `modem-stats`
`PING_TARGETS`   | A comma seperated string of targets to ping | `1.1.1.1,8.8.8.8,bbc.co.uk`
`FETCH_INTERVAL` | The frequency (in seconds) to fetch stats.  | `10` (default `10`)

Environment variables must also be passed in for `modem-stats` to run.
Check the [configuration section below](#Configuration).

`docker-compose.yaml` example:
```yaml
---
version: "2.1"
services:
  modem-stats:
    image: msh100/modem-stats
    container_name: modem-stats
    environment:
      - INFLUX_URL=http://influxdb:8086
      - INFLUX_DB=modem-stats
      - PING_TARGETS=1.1.1.1,8.8.8.8,bbc.co.uk
      - ROUTER_TYPE=superhub3
      - ROUTER_IP=192.168.100.1
    restart: unless-stopped
```


### Telegraf

If you are already running Telegraf, it makes more sense to add an extra input
to collect data from your modem.

To do this, we utilise the `input.execd` plugin (which exists in upstream
Telegraf).
Telegraf will be able to trigger a statistics fetch and interpret the output to
pass it over to your Telegraf outputs.

To run within Telegraf, you should [obtain a binary](#Downloading) for your
architecture and make that binary accessible to Telegraf (this may involve
mounting the binary if you are running Docker).

The Telegraf configuration should then use the `execd` input to collect data
from modem stats.

We need to set the `--daemon` flag to instruct Modem Stats to listen for new
STDIN lines to trigger data gathering.

```
[[inputs.execd]]
  command = ["/modem-stats", "--daemon"]
  data_format = "influx"
  signal = "STDIN"
```

To pass [configuration](#Configuration), you need to start Telegraf with those
environment variables defined.
If this is not possible, you can create a wrapper script to set those variables
and call `modem-stats`, like the following:

```bash
#!/bin/bash
/modem-stats --modem=superhub3 --daemon
```


### Prometheus

Is it possible to expose a Prometheus exporter.
If `--port` is defined, a webserver will start and Prometheus metrics will be
accessible at `/metrics`.

```
$ /modem-stats --modem=superhub3 --port=9000
```


## Binaries

The output of this repository is ultimately a single static binary with zero
dependencies.


### Downloading

Upon push to main, this repository builds and pushed binaries.
These can be downloaded from:

 * [Linux amd64](https://b2.msh100.uk/file/modem-stats/modem-stats.x86)
 * [Linux ARM5](https://b2.msh100.uk/file/modem-stats/modem-stats.arm5)
 * [MacOS ARM64](https://b2.msh100.uk/file/modem-stats/modem-stats.macos-arm64)

More architectures can be added on request.

In most cases, these binaries will be sufficient.


### Building

If you would like to build the binaries yourself, you will require Go 1.16 (may
work with earlier versions, but this is untested).

```
go build -o modem-stats sh-stats/*.go
```

For other architectures, extra options will need to be provided.
[Refer to this blog port for more information](https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04).


## Configuration

The scripts need to know the modem type (`ROUTER_TYPE` or `--modem=`).
Additional information depends on the model.

**Virgin Media Superhub 3<br/>
Ziggo Connectbox:**
 * `ROUTER_TYPE=superhub3` or `--modem=superhub3`
 * `ROUTER_IP` or `--ip=x.x.x.x` (defaults to `192.168.100.1`)

**Virgin Media Superhub 4:**
> **:warning: Warning:** Despite this statistics parser being fully functional, after some time the Superhub 4 fails to provide valid statistics until the device is rebooted. This is not an issue with the parser, but is an issue with the Superhub itself. [Issue](https://github.com/msh100/modem-stats/issues/2).
 * `ROUTER_TYPE=superhub4` or `--modem=superhub4`
 * `ROUTER_IP` or `--ip=x.x.x.x` (defaults to `192.168.100.1`)

**Virgin Media Superhub 5:**
 * `ROUTER_TYPE=superhub5` or `--modem=superhub5`
 * `ROUTER_IP` or `--ip=x.x.x.x` (defaults to `192.168.100.1`)

**Com Hem WiFi Hub C2:**
(This is likely to work on any Sagemcom DOCSIS modem)
 * `ROUTER_TYPE=comhemc2` or `--modem=comhemc2`
 * `ROUTER_IP` or `--ip=x.x.x.x` (defaults to `192.168.10.1`)
 * `ROUTER_USER` or `--username=admin` (defaults to `admin`)
 * `ROUTER_PASS` or `--password=password` (defaults to `admin`)

**Sky Hub 2:**
 * `ROUTER_TYPE=skyhub2`
 * `ROUTER_IP` or `--ip=x.x.x.x` (defaults to `192.168.0.1`)
 * `ROUTER_USER` or `--username=admin` (defaults to `admin`)
 * `ROUTER_PASS` or `--password=password` (defaults to `sky`)

**Ubee UBC1318:**
 * `ROUTER_TYPE=ubee`
 * `ROUTER_IP` or `--ip=x.x.x.x` (defaults to `192.168.100.1`)

**Technicolor TC4400:**
 * `ROUTER_TYPE=tc4400`
 * `ROUTER_IP` or `--ip=x.x.x.x` (defaults to `192.168.100.1`)
 * `ROUTER_USER` or `--username=user` (defaults to `user`)
 * `ROUTER_PASS` or `--password=password` (defaults to `password`)


### Example Usage

```
$ ROUTER_TYPE=superhub3 ROUTER_IP=192.168.100.1 go run main.go
downstream,channel=3,id=10,modulation=QAM256,scheme=SC-QAM frequency=211000000,snr=403,power=71,prerserr=300,postrserr=0
downstream,channel=9,id=16,modulation=QAM256,scheme=SC-QAM frequency=259000000,snr=409,power=68,prerserr=72,postrserr=0
...
$ go run main.go --modem=superhub3 --ip=192.168.100.1
downstream,channel=3,id=10,modulation=QAM256,scheme=SC-QAM frequency=211000000,snr=403,power=71,prerserr=300,postrserr=0
downstream,channel=9,id=16,modulation=QAM256,scheme=SC-QAM frequency=259000000,snr=409,power=68,prerserr=72,postrserr=0
...
```


## Grafana

You can add [the Router Stats dashboard](https://grafana.com/grafana/dashboards/14209)
to your Grafana instance by adding dashboard ID `14209`.
