#!/bin/bash

set -euo pipefail

pushd setup

./setup.sh

# TODO: figure out how to run just the application again without restarting Vault
#./run.sh

popd