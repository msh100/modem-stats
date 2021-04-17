#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

: ${INFLUX_URL:?INFLUX_URL not defined (example "http://influxdb:8086")}
: ${INFLUX_DB:?INFLUX_DB not defined (example "modem-stats")}

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

cat /etc/template/telegraf.conf |\
    sed "s/_PING_TARGETS/${PING_TARGETS}/" >\
    /etc/telegraf.d/telegraf.conf

exec telegraf \
    --config "/etc/telegraf.d/telegraf.conf" \
    --config-directory "/etc/telegraf.d/"
