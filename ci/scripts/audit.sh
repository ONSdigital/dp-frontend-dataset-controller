#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-frontend-dataset-controller
  make audit
popd