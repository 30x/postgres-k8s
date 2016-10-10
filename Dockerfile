FROM postgres:9.5.4
#FROM thirtyx/transicator:0.0.1

COPY image-scripts/init.sh /docker-entrypoint-initdb.d/

RUN chmod +x /docker-entrypoint-initdb.d/init.sh

COPY image-scripts/docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

RUN mkdir -p /clusterutils

COPY image-scripts/setreplicas.sh /clusterutils

COPY image-scripts/testdb.sh /clusterutils

RUN chmod +x -R /clusterutils


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
