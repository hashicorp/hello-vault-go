#!/bin/bash

set -euo pipefail

pushd scripts

./setup.sh

./build.sh

./run.sh

popd