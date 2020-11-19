# Superhub Channel Parser

A simply bash script to read channel diagnostics information from the Virgin
Media Superhub and output it in the InfluxDB line protocol.

This script is intended to be used within a Telegraf instance.

It may seem insane to have bash parsing JSON with regex however this has been
done intentionally as to prevent calls to jq, a process that is extremely
costly on a Raspberry Pi Zero.


## Usage

In its simplest form, you can run this script directly.
The only dependencies being bash and curl.

`ROUTER_IP` can be defined if the Superhub is not addressable at
`192.168.100.1` (the default Superhub IP in modem only mode).

```
$ bash routerStatus.sh
downstream,channel=1,id=25 frequency=331000000,snr=4,power=66
downstream,channel=2,id=9 frequency=203000000,snr=4,power=73
downstream,channel=3,id=10 frequency=211000000,snr=4,power=74
...
```


### Within Telegraf

The only configuration that needs to be added as an input for Telegraf is the
following:

```
[[inputs.exec]]
  commands = ["bash /vm-stats/routerStatus.sh"]
  data_format = "influx"
```

This repository can be mounted as a volume within Docker if necessary to
achieve this.
