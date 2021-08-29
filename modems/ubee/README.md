# Ubee DOCSIS Modem


## Supported Modems

The Ubee UBC1318, a DOCSIS modem used by Ziggo NL.

It is unknown if this processor works on non Ziggo provided modems or other modems from Ubee.


## Fetching the Data

The Ubee modem exposes statistics over its web interface on a page which requires no authentication.
A string containing JSON is stored on this page and parsed to provide the statistics overview.

This can be accessed at:

```
http://$ROUTER_IP/htdocs/cm_info_connection.php
```

## Interpreting the Data

Within the HTML contents there are a few variables in inline Javascript which we care about.

 * `ds_modulation` - A map of modulation IDs to strings (used in downstream channels).
 * `us_modulation` - A map of modulation IDs to strings (used in upstream channels).
 * `ifType` - A map of interface schemes to strings (used on all channels).
 * `cm_conn_json` - A string of a JSON array of active channels in the DOCSIS bond.

Within this processor, it's assumed the modulation and type maps are unchanging and these values are not dynamically loaded.


### `cm_conn_json` JSON Object

The JSON object contains many named fields, all with string values (except the channel objects arrays) regardless of the type.

 * `cm_conn_ds_gourpObj` - An array of downstream channels
 * `cm_conn_us_gourpObj` - An array of upstream channels


#### Downstream

Each object in the array contains the following fields:

 * `ds_type` - The interface scheme, mapped in `ifType`.
 * `ds_id` - Channel ID.
 * `ds_freq` - Channel frequency in Hz.
 * `ds_width` - Channel width.
 * `ds_power` - Channel power in dBmV.
 * `ds_snr` - Signal to noise ratio in dB.
   Note on 3.0 channels this is multipled by 10 and provided as an int.
 * `ds_modulation` - Channel modulation, mapped in `ds_modulation`.
 * `ds_correct` - Corrected codeword count.
 * `ds_uncorrect` - Uncorrectable codewords count.


#### Upstream

 * `us_status` - Status of the channel (Locked or not 0/1).
 * `us_type` - The interface scheme, mapped in `ifType`.
 * `us_id` - Channel ID.
 * `us_freq` - Channel frequency in Hz.
 * `us_width` - Channel width.
 * `us_power` - Channel power in dBmV.
 * `us_modulation` - Channel modulation, mapped in `us_modulation`.
