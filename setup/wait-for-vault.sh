#!/bin/sh
echo "Waiting for vault to start"

while [ ! -f "${1}" ]; do sleep 3; done

echo "Starting ${2}"
exec "${2}"