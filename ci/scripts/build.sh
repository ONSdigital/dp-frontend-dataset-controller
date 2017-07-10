#!/bin/bash -eux

cwd=$(pwd)

export GOPATH=$cwd/go

pushd $GOPATH/src/github.com/ONSdigital/dp-frontend-dataset-controller
  make build && cp build/dp-frontend-dataset-controller $cwd/build
popd
