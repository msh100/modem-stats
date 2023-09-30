# Technicolor TC4400 Channel Processor

## Supported Modems

This processor is only known to work with the Technicolor TC4400.


## Fetching the Data

The Technicolor TC4400 exposes a webserver which returns data in tables.
All the channel statistics are returned on a single page which requires BASIC
authentication.

All the channel statistics information can be fetched from the
`/cmconnectionstatus.html` endpoint.


## Interpreting the Data

The page returned is formatted HTML.
The page contains four tables, in a predictable order.

 1. Overall status
 2. Downstream channels
 3. Upstream channels
 4. DHCP leases


### Downstream

The second table on the page returns all of the downstream channels.

The first row of this table is a header in a single cell.
The second row is made up of the column headers.
Every subsequent row is data, with the columns made up of:

1. Channel index
2. Channel ID
3. Lock status
4. Channel type
5. Bonding status
6. Centre frequency (in Hz)
7. Channel width (in Hz)
8. SNR/MER threshold (value in dB)
9. Power level (in dBmV)
10. Modulation/Profile ID
11. Unerrored codewords
12. Corrected codewords
13. Uncorrectable codewords


### Upstream

The third table on the page returns all of the upstream channels.

As with the downstream channels, the first row of this table is a header in a
single cell.
The second row is made up of the column headers.
Every subsequent row is data, with the columns made up of:

1. Channel index
2. Channel ID
3. Lock status
4. Channel type
5. Bonding status
6. Centre frequency (in Hz)
7. Channel width (in Hz)
8. Power level (in dBmV)
9. Modulation/Profile ID
