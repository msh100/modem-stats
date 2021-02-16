#!/bin/bash
# Take the router status from VM and produce a influx friendly version
# Superhub 4 version

#DATA=$(curl "http://${ROUTER_IP:-192.168.0.1}/php/ajaxGet_device_networkstatus_data.php")
DATA=$(cat sno-alternative.data)

fetchJsVar() {
    local var
    var=$1
    regex="${var} = '?([^;']+)'?;"
    [[ $DATA =~ $regex ]]
    echo "${BASH_REMATCH[1]}"
}

function echoc() {
    local input output
    for input in "${@}"; do
        output+="${input},"
    done
    echo "${output::-1}"
}

echo -n "config,config=downstream"
echoc "maxrate=$(($(fetchJsVar "js_downStreamMaxTrafficRate") * 8))" \
      "maxburst=$(fetchJsVar "js_downStreamMaxTrafficBurst")"

echo -n "config,config=upstream"
echoc "maxrate=$(($(fetchJsVar "js_upStreamMaxTrafficRate") * 8))" \
      "maxburst=$(fetchJsVar "js_upStreamMaxTrafficBurst")"

echo "$(fetchJsVar "js_downStreamChannel")" | jq -r '.[] | "downstream,channel=\(.[0]),id=\(.[0]) frequency=\(.[1]),snr=\(.[3]),power=\(.[2]),prerserr=\(.[7]),postrserr=\(.[8])"'

echo "$(fetchJsVar "js_downStreamChannel31")" | jq -r '.[] | "downstream,channel=\(.[0]),id=\(.[0]) frequency=\(.[1]),snr=\(.[3]),power=\(.[8]),prerserr=\(.[9]),postrserr=\(.[10])"'

echo "$(fetchJsVar "js_upStreamChannel")" | jq -r '.[] | "upstream,channel=\(.[0]),id=\(.[0]) frequency=\(.[1]),power=\(.[2])"'
