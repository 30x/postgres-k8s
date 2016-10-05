#!/bin/bash

#------
# This file is a duplicate of this file.
# https://github.com/docker-library/postgres/blob/1c0bc9d905d569fead777b9b8e3836e8af1c394c/9.5/docker-entrypoint.sh
# It does not correctly detect files in an EBS volume.  As a result I'm changing the test for the PG_VERSION file
#------

set -e

if [ "${1:0:1}" = '-' ]; then
	set -- postgres "$@"
fi

if [ "$1" = 'postgres' ]; then

	#Wait for the mount of PGDATA to occur
	if [ ! "$PGMOUNT" ]; then
		echo "FATAL: You must specify PGMOUNT in order to use this init script."
		echo "FATAL: This should be equal to the persistent disk mount directory"
		exit 1
	fi

	MOUNTINFO=""
	MOUNT=0

	until [ $MOUNT -eq 1 ]; do
		echo "INFO: Checking if $PGMOUNT has been mounted"
		MOUNTINFO=$(df -h | grep $PGMOUNT)
		MOUNT=$(echo $PGMOUNT|wc -l)
	done

	echo "INFO: $PGMOUNT has been mounted, continuing"
	echo "$MOUNTINFO"


	mkdir -p "$PGDATA"
	chmod 700 "$PGDATA"
	chown -R postgres "$PGDATA"

	chmod g+s /run/postgresql
	chown -R postgres /run/postgresql

	# look specifically for PG_VERSION, as it is expected in the DB dir
	if [ ! -f "$PGDATA/PG_VERSION" ]; then

		echo "INFO: Could not find file $PGDATA/PG_VERSION.  Initializing the system"

		echo "Current data in $PGMOUNT"

		ls -al $PGMOUNT

		eval "gosu postgres initdb $POSTGRES_INITDB_ARGS"

		# check password first so we can output the warning before postgres
		# messes it up
		if [ "$POSTGRES_PASSWORD" ]; then
			pass="PASSWORD '$POSTGRES_PASSWORD'"
			authMethod=md5
		else
			# The - option suppresses leading tabs but *not* spaces. :)
			cat >&2 <<-'EOWARN'
				****************************************************
				WARNING: No password has been set for the database.
				         This will allow anyone with access to the
				         Postgres port to access your database. In
				         Docker's default configuration, this is
				         effectively any other container on the same
				         system.

				         Use "-e POSTGRES_PASSWORD=password" to set
				         it in "docker run".
				****************************************************
			EOWARN

			pass=
			authMethod=trust
		fi

		{ echo; echo "host all all 0.0.0.0/0 $authMethod"; } >> "$PGDATA/pg_hba.conf"

		# internal start of server in order to allow set-up using psql-client
		# does not listen on external TCP/IP and waits until start finishes
		gosu postgres pg_ctl -D "$PGDATA" \
			-o "-c listen_addresses='localhost'" \
			-w start

		: ${POSTGRES_USER:=postgres}
		: ${POSTGRES_DB:=$POSTGRES_USER}
		export POSTGRES_USER POSTGRES_DB

		psql=( psql -v ON_ERROR_STOP=1 )

		if [ "$POSTGRES_DB" != 'postgres' ]; then
			"${psql[@]}" --username postgres <<-EOSQL
				CREATE DATABASE "$POSTGRES_DB" ;
			EOSQL
			echo
		fi

		if [ "$POSTGRES_USER" = 'postgres' ]; then
			op='ALTER'
		else
			op='CREATE'
		fi
		"${psql[@]}" --username postgres <<-EOSQL
			$op USER "$POSTGRES_USER" WITH SUPERUSER $pass ;
		EOSQL
		echo

		psql+=( --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" )

		echo
		for f in /docker-entrypoint-initdb.d/*; do
			case "$f" in
				*.sh)     echo "$0: running $f"; . "$f" ;;
				*.sql)    echo "$0: running $f"; "${psql[@]}" < "$f"; echo ;;
				*.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${psql[@]}"; echo ;;
				*)        echo "$0: ignoring $f" ;;
			esac
			echo
		done

		gosu postgres pg_ctl -D "$PGDATA" -m fast -w stop

		echo
		echo 'PostgreSQL init process complete; ready for start up.'
		echo
	fi

	exec gosu postgres "$@"
else
	echo "INFO: Found $PGDATA/PG_VERSION.  Restarting the system with existing configuration"
fi

exec "$@"
