#!/bin/bash
# Take the router status from VM and produce a influx friendly version

data=$(curl --silent http://${ROUTER_IP:-192.168.100.1}/getRouterStatus)

# Downstream Channels
mibbase="1.3.6.1.2.1.10.127.1.1.1.1"
channel=1
while [ true ]; do
    channelmib="${mibbase}.1.${channel}"
    frequencymib="${mibbase}.2.${channel}"
    snrmib="${mibbase}.4.${channel}"
    powermib="${mibbase}.6.${channel}"

    if [[ $data  =~ ""$channelmib\":\"(-?[0-9]+)\" ]]; then
        channelid=${BASH_REMATCH[1]}
        [[ $data  =~ ""$frequencymib\":\"(-?[0-9]+)\" ]]
        frequency=${BASH_REMATCH[1]}
        [[ $data  =~ ""$snrmib\":\"(-?[0-9]+)\" ]]
        snr=${BASH_REMATCH[1]}
        [[ $data  =~ ""$powermib\":\"(-?[0-9]+)\" ]]
        power=${BASH_REMATCH[1]}

        echo "downstream,channel=${channel},id=${channelid} frequency=${frequency},snr=${snr},power=${power}"
    else
        break
    fi

    channel=$((channel+1))
done

# Upstream Channels
mibbase="1.3.6.1.2.1.10.127.1.1.2.1"
powerbase="1.3.6.1.4.1.4491.2.1.20.1.2.1.1"
channel=1
while [ true ]; do
    channelmib="${mibbase}.1.${channel}"
    frequencymib="${mibbase}.2.${channel}"
    powermib="${powerbase}.${channel}"

    if [[ $data  =~ ""$channelmib\":\"(-?[0-9]+)\" ]]; then
        channelid=${BASH_REMATCH[1]}
        [[ $data  =~ ""$frequencymib\":\"(-?[0-9]+)\" ]]
        frequency=${BASH_REMATCH[1]}
        [[ $data  =~ ""$powermib\":\"(-?[0-9]+)\" ]]
        power=${BASH_REMATCH[1]}

        echo "upstream,channel=${channel},id=${channelid} frequency=${frequency},power=${power}"
    else
        break
    fi

    channel=$((channel+1))
done
