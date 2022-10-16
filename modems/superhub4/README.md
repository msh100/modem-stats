# Superhub 4 Channel Processor

## Supported Modems

The Superhub 4 is distributed to multiple countries with different branding by
Liberty Global.

This processor is known to work on:

 - Virgin Media Superhub 4

Expected to work, but untested:

 - Virgin Media Ireland Hub 2.0
 - Ziggo Connect box Giga
 - UPC Switzerland Giga Connect Box
 - UPC Poland Giga Connect Box
 - UPC Slovakia Giga Connect Box


## Fetching the Data

The Superhub 4 (and equivilent international variants) exposes an endpoint over
HTTP which is used to display statistics on the web interface of the router.
This page is visible without authentication and is constructed by making an
extra HTTP call to an endpoint which returns a JSON array with some statistics
values.

A simple HTTP GET to
`http://$ROUTER_IP/php/ajaxGet_device_networkstatus_data.php` will return this
document.

The document returned is a single array.
```json
[
    "402750000",
    "46200000",
...
```

## Interpreting the Data

The array provided is in a predictable order.

Values that are returned as arrays need to be JSON loaded as they are JSON
arrays stored as strings.

Strangely, when an array is empty, one element exists within it with empty
values.
For example on a connection with no 3.1 upstream channels, the 3.1 upstream
channels value appears like this:
```json
"[[\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\"]]"
```


### Array order

 0.
 1.
 2.
 3.
 4.
 5.
 6.
 7.
 8. DOCSIS Version
 9. Loaded firmware version
 10.
 11. Downstream maximum rate
 12. Downstream maximum burst
 13.
 14.
 15. Upstream maximum rate
 16. Upstream maximum burst
 17.
 18.
 19. Scheduling type
 20. Array of 3.0 downstream channels
 21. Array of 3.0 upstream channels
 22. Array of log entries
 23. Array of 3.1 downstream channels
 24. Array of 3.1 upstream channels
 25. Count of 3.0 upstream channels
 26. Count of 3.0 downstream channels
 27. Count of 3.1 downstream channels
 28. Count of 3.1 upstream channels
 29.

Blanks are values that are currently unknown.
They should be easy to determine but no effort has gone into this yet.


### Channel Arrays

#### 3.0 Downstream Channels

Example:
```json
["30","402750000","5.400002","40.366287","QAM256","Locked","40.366287","5","0"]
```

0. Channel ID
1. Frequency
2. Power level
3. SNR
4. Modulation
5. Locked Status
6. SNR again *Unsure as to why*
7. Pre RS Errors
8. Post RS Errors


#### 3.0 Upstream Channels

Example:
```json
["3","46200000","46.770599","5120 KSym/sec","64QAM","US_TYPE_STDMA","0","0","0","0"]
```

0. Channel ID
1. Frequency
2. Power Level
3.
4. Modulation
5.
6.
7.
8.
9.


#### 3.1 Downstream Channels

Example:
```json
["33","96","4K","1880","QAM4096","759","Locked","43","5.5","186002","0"]
```

0. Channel ID
1. Frequency
2.
3.
4. Modulation
5. Locked status
6.
7. SNR
8. Power level
9. Pre RS Errors
10. Post RS Errors


#### 3.1 Upstream Channels

Example:
```json
["14","10.0","37.0","2K","QAM8","OFDMA","200","53.9","6","0"]
```

0. Channel ID
1. Channel Width
2. Power Level
3. FFT Type
4. Modulation
5. Channel Type
6. Number of subcarriers
7. First active subcarrier (MHz) **Note:** UI says this is Hz, but this looks like an error
8. T3 Timeouts
9. T4 Timeouts
