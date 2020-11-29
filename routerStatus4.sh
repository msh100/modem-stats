#!/bin/bash
# Take the router status from VM and produce a influx friendly version
# Superhub 4 version

DATA=$(curl "http://${ROUTER_IP:-192.168.0.1}/php/ajaxGet_device_networkstatus_data.php")

echo "${DATA}" | jq -r '. | "config,config=downstream maxrate=\(.[11]),maxburst=\(.[12])\nconfig,config=upstream maxrate=\(.[15]),maxburst=\(.[16])"'
echo "${DATA}" | jq -r '.[20]' | jq -r '.[] | "downstream,channel=\(.[0]),id=\(.[0]) frequency=\(.[1]),snr=\(.[3]),power=\(.[2]),prerserr=0,postrserr=0"'
echo "${DATA}" | jq -r '.[21]' | jq -r '.[] | "upstream,channel=\(.[0]),id=\(.[0]) frequency=\(.[1]),power=\(.[2])"'

# There doesn't appear to be pre/post rs error counter per channel on SH4
