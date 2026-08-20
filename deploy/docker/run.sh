#!/bin/sh
mkdir -p /data
# Wait for deploy-resolved admin credentials when this stack requires them.
if [ "${ADMIN_SECRETS_REQUIRED:-0}" = "1" ]; then
  attempts=0
  while [ -z "${ADMIN_USERNAME:-}" ] || [ -z "${ADMIN_PASSWORD:-}" ]; do
    attempts=$((attempts + 1))
    if [ "$attempts" -ge "${ADMIN_SECRETS_RETRIES:-30}" ]; then
      echo "admin vault credentials unavailable after ${ADMIN_SECRETS_RETRIES:-30} attempts" >&2
      exit 1
    fi
    echo "waiting for deploy-resolved admin vault credentials (${attempts}/${ADMIN_SECRETS_RETRIES:-30})" >&2
    sleep "${ADMIN_SECRETS_INTERVAL:-2}"
  done
fi

# Run pending migrations (idempotent).
/divoene migrate up
exec /divoene server
