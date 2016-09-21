FROM postgres:9.6

COPY build/posgres-agent /docker-entrypoint-initdb.d/ 

COPY init.sh /docker-entrypoint-initdb.d/