#!/bin/bash

set -euo pipefail

CWD=${PWD}

export CGO_ENABLED=0

GO_FLAGS=${GO_FLAGS:-"-tags netgo"}
GO_CMD=${GO_CMD:-"build"}
VERBOSE=${VERBOSE:-}
BUILD_NAME="ping"

REPO_PATH="github.com/fristonio/ping"

GOBIN=$PWD go "${GO_CMD}" -o "${BUILD_NAME}" ${GO_FLAGS} "${REPO_PATH}/cmd"

echo "[*] Setting capabilities for the binary"
sudo setcap cap_net_raw=ep "${BUILD_NAME}"

echo "[*] Build Complete."
exit 0
