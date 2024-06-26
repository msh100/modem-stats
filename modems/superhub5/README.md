# Superhub 5 Channel Processor

## Supported Modems

This processor is only known to work with the Virgin Media Superhub 5.


## Fetching the Data

The Superhub 5 exposes a REST API on its webserver at `/rest/v1`.
There are 3 endpoints which interest us here:

 * `/rest/v1/cablemodem/downstream`
 * `/rest/v1/cablemodem/upstream`
 * `/rest/v1/cablemodem/serviceflows`

The Superhub 5 runs at `192.168.0.1` in router mode and `192.168.100.1` in
modem mode.

## Interpreting the Data

JSON data is returned from each GET request.


### Downstream

`.downstream.channels` contains an array of downstream channels.
Each channel is made up of:

 - `channelId` - Channel ID
 - `frequency` - Frequench in hertz
 - `power` - Power in dBmV
 - `modulation` - Channel modulation (map below)
 - `snr` - Signal to Noise ratio in dB
 - `rxMer` - Signal to Noise ratio in dB (used by DOCSIS 3.1 channels)
 - `correctedErrors` - Count of corrected codewords
 - `uncorrectedErrors` - Count of uncorrectable codewords
 - `lockStatus` - (Bool) Channel locked

For example:

```json
{
  "channelId": 25,
  "frequency": 331000000,
  "power": 4.4,
  "modulation": "qam_256",
  "snr": 41,
  "rxMer": 41,
  "correctedErrors": 18,
  "uncorrectedErrors": 22,
  "lockStatus": true
}
```

The returned value for power is 10x greater on DOCSIS 3.1 channels than on
DOCSIS 3.0 channels, therefore they need to be normalised.

**Note:** It has been noted that the corrected count is displayed as "Pre RS
errors" in the Superhub UI, and uncorrected is displayed as "post RS errors".
The number for post was higher than pre which didn't make sense and I assume
that the Superhub 5 displays this data incorrectly.


### Upstream

`.upstream.channels` is similar to that of downstream and contains an array of
upstream channels.
Each channel is made up of:

 - `channelId` - Channel ID
 - `frequency` - Frequench in hertz
 - `lockStatus` - (Bool) Channel locked
 - `power` - Power in dBmV
 - `modulation` - Channel modulation (map below)
 - `t1Timeout` - T1 Timeout count
 - `t2Timeout` - T2 Timeout count
 - `t3Timeout` - T3 Timeout count
 - `t4Timeout` - T4 Timeout count
 - `channelType` - Type of upstream channel

For example:

```json
{
  "channelId": 1,
  "frequency": 60300000,
  "lockStatus": true,
  "power": 44.3,
  "symbolRate": 5120,
  "modulation": "qam_64",
  "t1Timeout": 0,
  "t2Timeout": 0,
  "t3Timeout": 0,
  "t4Timeout": 0,
  "channelType": "atdma"
}
```

The returned value for power is 10x greater on DOCSIS 3.1 channels than on
DOCSIS 3.0 channels, therefore they need to be normalised.


### Service Flows

To get the upstream and downstream rate/burst, we need to read from an array
at `.serviceFlows`.
Each object will contain an object at `.serviceFlow`.

We are interested in the value where `direction == "downstream"` or
`direction == "upstream"`.

Example:

```json
{
  "serviceFlow": {
    "serviceFlowId": 17980,
    "direction": "downstream",
    "maxTrafficRate": 402500089,
    "maxTrafficBurst": 42600,
    "minReservedRate": 0,
    "maxConcatenatedBurst": 0,
    "scheduleType": "undefined"
  }
}
```

### Modulation Map

Modulation is mapped by `/common/js/networkstatus.js` in the following ways:

```javascript
var MODULATION_MAP = {
    qpsk        : "QPSK",
    qam_8       : "QAM 8",
    qam_16      : "QAM 16",
    qam_32      : "QAM 32",
    qam_64      : "QAM 64",
    qam_128     : "QAM 128",
    qam_256     : "QAM 256",
    other       : "rs46",
    unsupported : "c_st30",
    error       : "c_st30",
    unknown     : "c_cd04" // 'Unknown'
};
```
