#!/bin/sh
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


# (re)start application and its dependencies
docker compose down --volumes
docker compose up -d --build
