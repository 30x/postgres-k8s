FROM postgres:9.5.4

COPY build/posgres-agent /postgres-config/posgres-agent

RUN chmod +x /postgres-config/posgres-agent

COPY scripts/init.sh /docker-entrypoint-initdb.d/

RUN chmod +x /docker-entrypoint-initdb.d/init.sh
