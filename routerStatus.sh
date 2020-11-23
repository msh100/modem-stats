#!/bin/bash
# Take the router status from VM and produce a influx friendly version

data=$(curl --silent "http://${ROUTER_IP:-192.168.100.1}/getRouterStatus")

function contains() {
    local n=$#
    local value=${!n}
    for ((i=1;i < $#;i++)) {
        if [ "${!i}" == "${value}" ]; then
            return 0
        fi
    }
    return 1
}

# Active bonded channels
channelmib="1.3.6.1.4.1.4115.1.3.4.1.1.12.0"
[[ $data  =~ \"$channelmib\":\"([0-9,]+)\" ]]
IFS=', ' read -r -a activechannels <<< "${BASH_REMATCH[1]}"

downmibbase="1.3.6.1.2.1.10.127.1.1.1.1"
upmibbase="1.3.6.1.2.1.10.127.1.1.2.1"
snrbase="1.3.6.1.2.1.10.127.1.1.4.1.5"
bondbase="1.3.6.1.2.1.10.127.1.1.4.1"
powerbase="1.3.6.1.4.1.4491.2.1.20.1.2.1.1"
downchannels=()

for channelid in "${activechannels[@]}"; do
    # Downstream channels
    if [[ $data  =~ \"$downmibbase.1.([0-9]+)\":\"$channelid\" ]] && \
       ! contains "${downchannels[@]}" "${BASH_REMATCH[1]}"; then
        channel=${BASH_REMATCH[1]}
        frequencymib="${downmibbase}.2.${channel}"
        snrmib="${snrbase}.${channel}"
        powermib="${downmibbase}.6.${channel}"
        prerserrmib="${bondbase}.3.${channel}"
        postrserrmib="${bondbase}.4.${channel}"
        downchannels+=("${channel}")

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
    # Upstream channels
    elif [[ $data  =~ \"$upmibbase.1.([0-9]+)\":\"$channelid\" ]]; then
        channel=${BASH_REMATCH[1]}
        frequencymib="${upmibbase}.2.${channel}"
        powermib="${powerbase}.${channel}"

        [[ $data  =~ \"$frequencymib\":\"(-?[0-9]+)\" ]]
        frequency=${BASH_REMATCH[1]}
        [[ $data  =~ \"$powermib\":\"(-?[0-9]+)\" ]]
        power=${BASH_REMATCH[1]}

        echo -n "upstream,channel=${channel},id=${channelid} "
        echo \
            "frequency=${frequency}" \
            "power=${power}" \
            | tr ' ' ','
    fi
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
