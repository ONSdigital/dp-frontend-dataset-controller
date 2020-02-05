#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-frontend-dataset-controller
  make build && cp build/dp-frontend-dataset-controller $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
