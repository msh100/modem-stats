#!/bin/bash
# Take the router status from VM and produce a influx friendly version

DATA=$(curl --silent "http://${ROUTER_IP:-192.168.100.1}/getRouterStatus")

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

function read_mib() {
    [[ "${DATA}" =~ \"${1}\":\"([^\"]+)\",? ]] && echo "${BASH_REMATCH[1]}"
}

function echoc() {
    local input output
    for input in "${@}"; do
        output+="${input},"
    done
    echo "${output::-1}"
}

# Active bonded channels
channelmib="1.3.6.1.4.1.4115.1.3.4.1.1.12.0"
IFS=', ' read -r -a activechannels <<< "$(read_mib "${channelmib}")"

downmibbase="1.3.6.1.2.1.10.127.1.1.1.1"
upmibbase="1.3.6.1.2.1.10.127.1.1.2.1"
snrbase="1.3.6.1.2.1.10.127.1.1.4.1.5"
bondbase="1.3.6.1.2.1.10.127.1.1.4.1"
powerbase="1.3.6.1.4.1.4491.2.1.20.1.2.1.1"
downchannels=()

for channelid in "${activechannels[@]}"; do
    # Downstream channels
    if [[ "${DATA}" =~ \"$downmibbase.1.([0-9]+)\":\"$channelid\" ]] && \
       ! contains "${downchannels[@]}" "${BASH_REMATCH[1]}"; then
        channel=${BASH_REMATCH[1]}
        downchannels+=("${channel}")

        echo -n "downstream,channel=${channel},id=${channelid} "
        echoc \
            "frequency=$(read_mib "${downmibbase}.2.${channel}")" \
            "snr=$(read_mib "${snrbase}.${channel}")" \
            "power=$(read_mib "${downmibbase}.6.${channel}")" \
            "prerserr=$(read_mib "${bondbase}.3.${channel}")" \
            "postrserr=$(read_mib "${bondbase}.4.${channel}")"
    # Upstream channels
    elif [[ "${DATA}" =~ \"$upmibbase.1.([0-9]+)\":\"$channelid\" ]]; then
        channel=${BASH_REMATCH[1]}

        echo -n "upstream,channel=${channel},id=${channelid} "
        echoc \
            "frequency=$(read_mib "${upmibbase}.2.${channel}")" \
            "power=$(read_mib "${powerbase}.${channel}")"
    fi
done

# Determine config
upstreamid="1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.1."
downstreamid="1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.2."
ratemib="1.3.6.1.4.1.4491.2.1.21.1.2.1.6.2.1"
burstmib="1.3.6.1.4.1.4491.2.1.21.1.2.1.7.2.1"

if [[ "${DATA}" =~ \"$upstreamid([0-9]+)\":\"1\" ]]; then
    upstreamid=${BASH_REMATCH[1]}

    echo -n "config,config=upstream "
    echoc \
        "maxrate=$(read_mib "${ratemib}.${upstreamid}")" \
        "maxburst=$(read_mib "${burstmib}.${upstreamid}")"
fi

if [[ "${DATA}" =~ \"$downstreamid([0-9]+)\":\"1\" ]]; then
    downstreamid=${BASH_REMATCH[1]}

    echo -n "config,config=downstream "
    echoc \
        "maxrate=$(read_mib "${ratemib}.${downstreamid}")" \
        "maxburst=$(read_mib "${burstmib}.${downstreamid}")"
fi
