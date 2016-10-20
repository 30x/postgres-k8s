#FROM postgres:9.5.4
FROM thirtyx/transicator-postgres:0.0.2

#Remove scripts installed for dev purposes.  We want our utilities to configure the images
RUN rm -rf /docker-entrypoint-initdb.d/*

COPY init.sh /docker-entrypoint-initdb.d/

RUN chmod +x /docker-entrypoint-initdb.d/init.sh

#We don't need this.  This is the default script with extra debugging
#COPY docker-entrypoint.sh /docker-entrypoint.sh
#RUN chmod +x /docker-entrypoint.sh

COPY clusterutils /clusterutils

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
