#!/bin/sh

./bin/ekanek \
-db-host "${DB_HOST}" \
-db-name "${DB_NAME}" \
-db-pass "${DB_PASS}" \
-db-port "${DB_PORT}" \
-db-user "${DB_USER}" \
-log-level "${LOG_LEVEL}" \
-srv-timeout "${SRV_TIMEOUT}"
