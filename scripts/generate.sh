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
mockery --dir ${project_root}/pkg --recursive --keeptree --output ${mock_dir} --name '^(Server|Client|Database|Tx|Checker)$'

go_redis=$(go list -m -f "{{ .Dir }}" github.com/go-redis/redis)
mockery --dir ${go_redis} --output ${mock_dir}/go-redis --name '^Cmdable|(Status|String|Int)Cmd$'
echo "Generated mocks"
