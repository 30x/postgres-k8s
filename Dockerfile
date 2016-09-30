FROM postgres:9.5.4

COPY scripts/init.sh /docker-entrypoint-initdb.d/

RUN chmod +x /docker-entrypoint-initdb.d/init.sh




# -------------
# This is for debugging purposes only.  This will start the container then block,
# allowing the user to interactively launch a bash shell into the container to run and debug
# -------------

COPY scripts/forever.sh /

RUN chmod +x /forever.sh

ENTRYPOINT ["/forever.sh"]
