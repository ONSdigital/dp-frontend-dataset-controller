#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-frontend-dataset-controller
  make lint
popd
