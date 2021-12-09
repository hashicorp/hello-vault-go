#!/bin/sh
echo "Waiting for orchestrator"

while [ ! -f /tmp/secret ]; do sleep 3; done

sleep 2;
echo "Starting ${1}"

export VAULT_APPROLE_ROLE_ID=$(cat /tmp/role)

exec "${1}"
