//go:build tools

// This file will never be built, but `go mod tidy` will see the packages
// imported here as dependencies and not remove them from `go.mod`.

package main

import (
	_ "github.com/go-jose/go-jose/v4"
	_ "github.com/golang/glog"
	_ "github.com/kevinburke/go-bindata"
	_ "golang.org/x/crypto/hkdf"
)
