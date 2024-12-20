#!/bin/bash

OUTPUT_DIR=$PWD/dist
mkdir -p "${OUTPUT_DIR}"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OUTPUT_DIR}/panzer-plugin_linux_amd64 -ldflags "-X github.com/metskem/panzer-plugin/version.VERSION=${VERSION} -X github.com/metskem/panzer-plugin/version.COMMIT=${COMMIT}" .
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${OUTPUT_DIR}/panzer-plugin_darwin_amd64 -ldflags "-X github.com/metskem/panzer-plugin/version.VERSION=${VERSION} -X github.com/metskem/panzer-plugin/version.COMMIT=${COMMIT}" .
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ${OUTPUT_DIR}/panzer-plugin_darwin_arm64 -ldflags "-X github.com/metskem/panzer-plugin/version.VERSION=${VERSION} -X github.com/metskem/panzer-plugin/version.COMMIT=${COMMIT}" .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${OUTPUT_DIR}/panzer-plugin_windows_amd64 -ldflags "-X github.com/metskem/panzer-plugin/version.VERSION=${VERSION} -X github.com/metskem/panzer-plugin/version.COMMIT=${COMMIT}" .
