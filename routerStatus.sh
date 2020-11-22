#!/bin/bash
# Take the router status from VM and produce a influx friendly version

data=$(curl --silent http://${ROUTER_IP:-192.168.100.1}/getRouterStatus)

# Downstream Channels
mibbase="1.3.6.1.2.1.10.127.1.1.1.1"
snrbase="1.3.6.1.2.1.10.127.1.1.4.1.5"
bondbase="1.3.6.1.2.1.10.127.1.1.4.1"

channel=1
while [ true ]; do
    channelmib="${mibbase}.1.${channel}"
    frequencymib="${mibbase}.2.${channel}"
    snrmib="${snrbase}.${channel}"
    powermib="${mibbase}.6.${channel}"
    prerserrmib="${bondbase}.3.${channel}"
    postrserrmib="${bondbase}.4.${channel}"

    if [[ $data  =~ \"$channelmib\":\"(-?[0-9]+)\" ]]; then
        channelid=${BASH_REMATCH[1]}
        [[ $data  =~ \"$frequencymib\":\"(-?[0-9]+)\" ]]
        frequency=${BASH_REMATCH[1]}
        [[ $data  =~ \"$snrmib\":\"(-?[0-9]+)\" ]]
        snr=${BASH_REMATCH[1]}
        [[ $data  =~ \"$powermib\":\"(-?[0-9]+)\" ]]
        power=${BASH_REMATCH[1]}
        [[ $data  =~ \"$prerserrmib\":\"(-?[0-9]+)\" ]]
        prerserr=${BASH_REMATCH[1]}
        [[ $data  =~ \"$postrserrmib\":\"(-?[0-9]+)\" ]]
        postrserr=${BASH_REMATCH[1]}

        echo -n "downstream,channel=${channel},id=${channelid} "
        echo \
            "frequency=${frequency}" \
            "snr=${snr}" \
            "power=${power}" \
            "prerserr=${prerserr}" \
            "postrserr=${postrserr}" \
            | tr ' ' ','
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

    if [[ $data  =~ \"$channelmib\":\"(-?[0-9]+)\" ]]; then
        channelid=${BASH_REMATCH[1]}
        [[ $data  =~ \"$frequencymib\":\"(-?[0-9]+)\" ]]
        frequency=${BASH_REMATCH[1]}
        [[ $data  =~ \"$powermib\":\"(-?[0-9]+)\" ]]
        power=${BASH_REMATCH[1]}

        echo -n "upstream,channel=${channel},id=${channelid} "
        echo \
            "frequency=${frequency}" \
            "power=${power}" \
            | tr ' ' ','
    else
        break
    fi

    channel=$((channel+1))
done

# Determine config
upstreamid="1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.1."
downstreamid="1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.2."
ratemib="1.3.6.1.4.1.4491.2.1.21.1.2.1.6.2.1"
burstmib="1.3.6.1.4.1.4491.2.1.21.1.2.1.7.2.1"

if [[ $data  =~ \"$upstreamid([0-9]+)\":\"1\" ]]; then
    upstreamid=${BASH_REMATCH[1]}

    [[ $data  =~ \"$ratemib.$upstreamid\":\"([0-9]+)\" ]]
    maxrate=${BASH_REMATCH[1]}
    [[ $data  =~ \"$burstmib.$upstreamid\":\"([0-9]+)\" ]]
    maxburst=${BASH_REMATCH[1]}

    echo "config,config=upstream maxrate=${maxrate},maxburst=${maxburst}"
fi

if [[ $data  =~ \"$downstreamid([0-9]+)\":\"1\" ]]; then
    downstreamid=${BASH_REMATCH[1]}

    [[ $data  =~ \"$ratemib.$downstreamid\":\"([0-9]+)\" ]]
    maxrate=${BASH_REMATCH[1]}
    [[ $data  =~ \"$burstmib.$downstreamid\":\"([0-9]+)\" ]]
    maxburst=${BASH_REMATCH[1]}

    echo "config,config=downstream maxrate=${maxrate},maxburst=${maxburst}"
fi
