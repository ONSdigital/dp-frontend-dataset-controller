#!/bin/bash -eux

export GOPATH=$(pwd)

pushd $GOPATH/dp-frontend-dataset-controller
  make test
popd
