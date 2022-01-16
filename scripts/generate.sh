#/usr/bin/env bash

set -e

trap 'catch $?' EXIT

catch () {
    if [[ "$1" != "0" ]]; then
        echo "Failed"
        exit $1
    fi
}

project_root=$(cd $(dirname $0)../ >/dev/null 2>&1; pwd)

mock_dir=${project_root}/internal/mocks
mockery --dir ${project_root}/pkg --recursive --keeptree --output ${mock_dir} --name '^Server$|^Client$|Database|Tx|Checker'

echo "Generated mocks"
