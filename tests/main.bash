#!/usr/bin/env bash

set -euo pipefail

export SPLITTER_LOG_LEVEL=info

splitter \
  deploy \
  -n case1 \
  -f .fixtures/app.apk

splitter \
  --format raw \
  deploy \
  -n case2 \
  -f .fixtures/app.apk

splitter \
  --format markdown \
  deploy \
  -n case3 \
  -f .fixtures/app.apk

splitter \
  --config ./splitter.another.yml \
  deploy \
  -n case4 \
  -f .fixtures/app.apk
