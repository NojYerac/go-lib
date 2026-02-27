#/usr/bin/env bash

set -e

trap 'catch $?' EXIT

catch () {
    if [[ "$1" != "0" ]]; then
        echo "Failed"
        exit $1
    fi
}

ginkgo -r
golangci-lint run
mockery
echo "Generated mocks"
