#!/usr/bin/env bash

set -euo pipefail

export SPLITTER_LOG_LEVEL=info

splitter \
  distribute \
  -n case1 \
  -f .fixtures/app.apk

splitter \
  --format raw \
  distribute \
  -n case2 \
  -f .fixtures/app.apk

splitter \
  --format markdown \
  distribute \
  -n case3 \
  -f .fixtures/app.apk

splitter \
  --config ./splitter.another.yml \
  distribute \
  -n case4 \
  -f .fixtures/app.apk
