#!/bin/sh

endpoint_url=http://s3-fake:4572/

sleep 10

aws --endpoint-url=$endpoint_url s3 mb s3://ekanek

sleep 5

migrate -database postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE} -path migrations/ up

./bin/ekanek \
-db-host "${DB_HOST}" \
-db-name "${DB_NAME}" \
-db-pass "${DB_PASS}" \
-db-port "${DB_PORT}" \
-db-user "${DB_USER}" \
-log-level "${LOG_LEVEL}" \
-srv-timeout "${SRV_TIMEOUT}" \
-aws-region "${AWS_DEFAULT_REGION}" \
-aws-key "${AWS_ACCESS_KEY_ID}" \
-aws-secret "${AWS_SECRET_ACCESS_KEY}" \
-private-key "${PRIVATE_KEY}"
