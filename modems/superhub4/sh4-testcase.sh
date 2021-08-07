#!/bin/bash

# Superhub 4 statistics failure test case.
# This simple script is designed to demonstrate that
# https://github.com/msh100/modem-stats/issues/2 affects all Superhub 4 users.

# This script will do the following:
#  1) Ensure the Superhub 4 is in a bad state
#  2) Wait until the Superhub 4 is healthy (rebooted since step 1).
#  3) Query the Superhub 4 until it is in a bad state again (10 consecutive failures).
#  4) Print time and number of queries taken to break the Superhub.

is_broken() {
    local broken_stats data
    data="$1"
    broken_stats='["","","","",null,"","","","","","","","","","","","","","","","[[\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\"]]","[[\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\"]]","[]","[[\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\"]]","[[\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\"]]","","","","",""]'

    [ "${data}" == "${broken_stats}" ]
}

get_stats_from_superhub() {
    local router_ip data
    router_ip=$1
    query_address="http://${router_ip}/php/ajaxGet_device_networkstatus_data.php"
    curl -s "${query_address}"
}

to_log() {
    echo "[$(date)]" "${@}"
}

main() {
    local query
    local router_ip
    router_ip="$1"
    query="$(get_stats_from_superhub "${router_ip}")"

    if ! is_broken "${query}"; then
        to_log "The Superhub 4 needs to be in a bad state before starting this script"
        exit 1
    fi

    to_log "Superhub 4 is in a bad state. Stop all queries to the router and then restart the router to begin the test."

    until ! is_broken "${query}"; do
        query="$(get_stats_from_superhub "${router_ip}")"
        sleep 10
    done

    to_log "Superhub 4 is healthy, starting test."

    local count broken_count
    count=0
    broken_count=0
    until [ "${broken_count}" == "10" ]; do
        ((count=count+1))
        query="$(get_stats_from_superhub "${router_ip}")"

        if is_broken "${query}"; then
            ((broken_count=broken_count+1))
        else
            broken_count=0
        fi

        sleep 10
    done

    to_log "Superhub 4 statistics are now broken after ${count} queries"
}

main "$@"
