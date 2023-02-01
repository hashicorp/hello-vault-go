#!/bin/sh
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


APP_ADDRESS="http://localhost:8080"

# bring up hello-vault-go service and its dependencies
docker compose up -d --build --quiet-pull

# bring down the services on exit
trap 'docker compose down --volumes' EXIT

# TEST 1: POST /payments (static secrets)
output1=$(curl --silent --request POST "${APP_ADDRESS}/payments")

echo "[TEST 1]: output: $output1"

if [ "${output1}" != '{"message":"hello world!"}' ]
then
    echo "[TEST 1]: FAILED: unexpected output"
    exit 1
else
    echo "[TEST 1]: OK"
fi

# TEST 2: GET /products (dynamic secrets)
output2=$(curl --silent --request GET "${APP_ADDRESS}/products")

echo "[TEST 2]: output: $output2"

if [ "${output2}" != '[{"id":1,"name":"Rustic Webcam"},{"id":2,"name":"Haunted Coloring Book"}]' ]
then
    echo "[TEST 2]: FAILED: unexpected output"
    exit 1
else
    echo "[TEST 2]: OK"
fi
