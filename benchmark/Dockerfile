FROM postgres:9.5.4

RUN apt-get update && \
  apt-get install curl python groff gnuplot -y &&  \
  curl -O https://bootstrap.pypa.io/get-pip.py && \
  python get-pip.py && \
  pip install awscli

COPY pgbench /pgbench

RUN chmod -R 755 /pgbench

CMD /bin/bash -c 'cd /pgbench && ./runtests.sh'
