FROM postgres:9.5.4


COPY init.sh /docker-entrypoint-initdb.d/

COPY master-node.sh /nodesetup/

COPY replica-node.sh /nodesetup/

COPY slave-node.sh /nodesetup/

COPY util.sh /nodesetup/
