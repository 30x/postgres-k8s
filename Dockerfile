FROM postgres:9.5.4
#FROM thirtyx/transicator:0.0.1

COPY image-scripts/init.sh /docker-entrypoint-initdb.d/

RUN chmod +x /docker-entrypoint-initdb.d/init.sh

COPY image-scripts/docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

COPY image-scripts/setreplicas.sh /

RUN chmod +x /setreplicas.sh


# -------------
# This is for debugging purposes only.  This will start the container then block,
# allowing the user to interactively launch a bash shell into the container to run and debug and configure start and stop PG without a container reboot
# -------------
#
#COPY scripts/forever.sh /
#
#RUN chmod +x /forever.sh
#
#ENTRYPOINT ["/forever.sh"]
