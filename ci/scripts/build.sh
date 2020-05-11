#!/bin/bash -eux

export BINPATH=$(pwd)/build
export cwd=$(pwd)

pushd $cwd/dp-frontend-dataset-controller
  BINPATH=$BINPATH make build
  cp Dockerfile.concourse $BINPATH
popd