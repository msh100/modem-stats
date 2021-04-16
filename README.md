# Modem Statistics Parser

A utility to read channel diagnostics information from DOCSIS and VDSL modems
and output it in the InfluxDB line protocol.

This package is intended to be used within a Telegraf instance.

This package has been written in Go in an attempt to allow it to run on low end
hardware (such as a Raspberry Pi Zero) with no issues.

![Grafana dashboard screenshot](https://user-images.githubusercontent.com/4477262/114266746-cd910980-99ef-11eb-8a5e-f4f719897719.JPG)


## Usage

In its simplest form, you can run this repository directly.
The only dependency being Go.

A compiled binary of this repository will require no dependencies.


### Within Docker

A Docker image exists for this repository with Telegraf configured to write to
InfluxDB at `msh100/modem-stats`.
This image currently only supports X86.

The following environment variables must be set:

Name           | Description                                 | Example
---------------|---------------------------------------------|---------------------
`INFLUX_URL`   | The HTTP API URI for your InfluxDB server.  | `http://influxdb:8086`
`INFLUX_DB`    | The InfluxDB database to use                | `modem-stats`
`PING_TARGETS` | A comma seperated string of targets to ping | `1.1.1.1,8.8.8.8,bbc.co.uk`

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
      - IP_ADDRESS=192.168.100.1
    restart: unless-stopped
```


### Downloading Binaries

Binaries for this repository can be fetched from:

 * [amd64](https://b2.msh100.uk/file/modem-stats/modem-stats.x86)
 * [ARM5](https://b2.msh100.uk/file/modem-stats/modem-stats.arm5)

More architectures can be added on request.


### Configuration

The scripts need to know the modem type (`ROUTER_TYPE`).
Additional information depends on the model.

**Virgin Superhub 3<br/>
Ziggo Connectbox:**
 * `ROUTER_TYPE=superhub3`
 * `IP_ADDRESS` (defaults to `192.168.100.1`)

**Virgin Superhub 4:**
> **:warning: Warning:** Despite this statistics parser being fully functional, after some time the Superhub 4 fails to provide valid statistics until the device is rebooted. This is not an issue with the parser, but is an issue with the Superhub itself. [Issue](https://github.com/msh100/modem-stats/issues/2).
 * `ROUTER_TYPE=superhub4`
 * `IP_ADDRESS` (defaults to `192.168.100.1`)

**Com Hem WiFi Hub C2:**
(This is likely to work on any Sagemcom DOCSIS modem)
 * `ROUTER_TYPE=comhemc2`
 * `IP_ADDRESS` (defaults to `192.168.10.1`)
 * `ROUTER_USER` (defaults to `admin`)
 * `ROUTER_PASS` (defaults to `admin`)

**Sky Hub 2:**
 * `ROUTER_TYPE=skyhub2`
 * `IP_ADDRESS` (defaults to `192.168.0.1`)
 * `ROUTER_USER` (defaults to `admin`)
 * `ROUTER_PASS` (defaults to `sky`)

### Example

```
$ ROUTER_TYPE=superhub3 IP_ADDRESS=192.168.100.1 go run sh-stats/*.go
downstream,channel=3,id=10,modulation=QAM256,scheme=SC-QAM frequency=211000000,snr=403,power=71,prerserr=300,postrserr=0
downstream,channel=9,id=16,modulation=QAM256,scheme=SC-QAM frequency=259000000,snr=409,power=68,prerserr=72,postrserr=0
...
```

### Within Telegraf

To run within Telegraf, you should build a binary for your architecture, then
mount that executable to the container.

```
go build -o modem-stats sh-stats/*.go
```

`modem-stats` should be mounted to the container.

The Telegraf configuration should then use the `exec` input to call it.

```
[[inputs.exec]]
  commands = ["/modem-stats"]
  data_format = "influx"
```

## Grafana

You can add [the Router Stats dashboard](https://grafana.com/grafana/dashboards/14209)
to your Grafana instance by adding dashboard ID `14209`.
