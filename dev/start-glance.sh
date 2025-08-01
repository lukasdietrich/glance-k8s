#!/usr/bin/env sh

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"

set -e
set -x

docker run \
	--pull=always \
	--rm \
	--network=host \
	--volume "${SCRIPT_DIR}/glance.yml:/app/config/glance.yml:ro" \
	"glanceapp/glance:latest"
