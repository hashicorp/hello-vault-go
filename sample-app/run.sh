#!/bin/sh
# Copyright IBM Corp. 2021, 2026
# SPDX-License-Identifier: MPL-2.0


# (re)start application and its dependencies
docker compose down --volumes
docker compose up -d --build
