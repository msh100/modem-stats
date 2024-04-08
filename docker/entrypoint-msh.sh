#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

: ${INFLUX_URL:?INFLUX_URL not defined (example "http://influxdb:8086")}
: ${INFLUX_DB:?INFLUX_DB not defined (example "modem-stats")}
FETCH_INTERVAL="${FETCH_INTERVAL:-10}"

# Fetch interval needs to be a numerical value
if ! [[ "${FETCH_INTERVAL}" =~ ^[0-9]+$ ]] ; then
   echo "FETCH_INTERVAL must be numerical" >&2
   exit 1
fi

# Create a ping targets array from comma seperated string of targets
PING_TARGETS="${PING_TARGETS:-""}"
if [ "${PING_TARGETS}" != "" ]; then
    PING_TARGETS_ARR=(${PING_TARGETS//,/ })
    PING_TARGETS_STRING=""
    for TARGET_VALUE in "${PING_TARGETS_ARR[@]}"; do
        PING_TARGETS_STRING="${PING_TARGETS_STRING}\"${TARGET_VALUE}\", "
    done
    PING_TARGETS="[ ${PING_TARGETS_STRING::-2} ]"
else
    PING_TARGETS="[]"
fi

# We can't template an array, so need to write this before exec
cat /etc/template/telegraf.conf |\
    sed "s/_PING_TARGETS/${PING_TARGETS}/" >\
    /etc/telegraf.d/telegraf.conf

export FETCH_INTERVAL
export INFLUX_URL
export INFLUX_DB

exec telegraf \
    --config "/etc/telegraf.d/telegraf.conf" \
    --config-directory "/etc/telegraf.d/"
