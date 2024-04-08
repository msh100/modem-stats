FROM scratch

ADD "output/modem-stats${OBJECT_SUFFIX}.x86" /modem-stats

ENTRYPOINT /modem-stats --port=9090
