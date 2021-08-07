# Superhub 3 Channel Processor

## Supported Modems

The Superhub 3 is distributed to multiple countries with different branding by
Liberty Global.

This processor is known to work on:

 - Virgin Media Superhub 3
 - Ziggo Connectbox

Expected to work, but untested:

 - Virgin Media Ireland Superhub 3
 - UPC Switzerland Connect Box
 - UPC Poland Connect Box
 - UPC Slovakia prémiový modem
 - Unity Media Germany Connect Box


## Fetching the Data

The Superhub 3 (and equivilent international variants) exposes an endpoint over
HTTP which is used to display statistics on the web interface of the router.
This page is visible without authentication and is constructed by making an
extra HTTP call to an endpoint which returns a JSON document with SNMP MIBs and
their associated values.

A simple HTTP GET to `http://$ROUTER_IP/getRouterStatus` will return this
document.

The document has the following structure as a single dimentional object:
```json
{
    "MIB1": "VALUE",
    "MIB2": "VALUE"
}
```

## Interpreting the Data


### Profile Speeds

Initially we search for the profiles that represent both the upsteam and
downstream speeds.

```
1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.$ID1.$ID2
```

We take MIBs which have this structure and have a value of `1`.
The highest odd number of ID1 in this result set corresponds to uplink, and the
highest even number of ID1 corresponds to downlink.
ID2 is the value we need.

```
1.3.6.1.4.1.4491.2.1.21.1.2.1.6.2.1.$ID2 = MaxRate
1.3.6.1.4.1.4491.2.1.21.1.2.1.7.2.1.$ID2 = MaxBurst
```


### Channel Information

Each channel has a number of values.
Note that the channel ID and what the Superhub considers to be the ID are
not the same values.

The MIBs that are used for **downstream** channels are:

```
1.3.6.1.2.1.10.127.1.1.1.1.$ID = Channel ID
1.3.6.1.2.1.10.127.1.1.1.1.2.$ID = Frequency
1.3.6.1.2.1.10.127.1.1.1.1.6.$ID = Power Level
1.3.6.1.2.1.10.127.1.1.4.1.5.$ID = SNR
1.3.6.1.2.1.10.127.1.1.4.1.3.$ID = Pre RS Errors
1.3.6.1.2.1.10.127.1.1.4.1.4.$ID = Post RS Errors
```

The MIBs that are used for **upstream** channels are:

```
1.3.6.1.2.1.10.127.1.1.2.1.1.$ID = Channel ID
1.3.6.1.2.1.10.127.1.1.2.1.1.$ID = Frequency
1.3.6.1.4.1.4491.2.1.20.1.2.1.1.$ID = Power Level
```


#### Detemine Channels

The method we use to determine channels is by iterating over MIB strings to
find matches.

```
1.3.6.1.2.1.10.127.1.1.1.1.1.([0-9]+) = Upstream Channels
1.3.6.1.2.1.10.127.1.1.2.1.1.([0-9]+) = Downstream Channels
```

There's also a MIB that lists all channels:

```
1.3.6.1.4.1.4115.1.3.4.1.1.12.0
```

The issue we encountered here is the unpredictability in switching from
downstream to upstream channels.
