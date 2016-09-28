FROM postgres:9.5.4

COPY scripts/init.sh /docker-entrypoint-initdb.d/

RUN chmod +x /docker-entrypoint-initdb.d/init.sh
