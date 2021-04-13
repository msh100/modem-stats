FROM telegraf:1.16.2 as builder

ADD modem-stats.x86 /modem-stats
RUN chmod +x /modem-stats

ADD ./entrypoint-msh.sh /entrypoint-msh.sh
RUN chmod +x /entrypoint-msh.sh

RUN mkdir -p /etc/telegraf.d/
ADD ./telegraf.conf /etc/telegraf.d/

# We build from scratch as to remove all the volume and exposures from the
# source Telegraf Docker image. Since we don't have any state or listeners,
# this is "okay" to do.
FROM scratch

COPY --from=builder / /

ENTRYPOINT /entrypoint-msh.sh
