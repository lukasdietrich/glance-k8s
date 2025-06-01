#!/usr/bin/env sh

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"

set -e
set -x

podman run \
	--pull=newer \
	--rm \
	--network=host \
	--volume "${SCRIPT_DIR}/glance.yml:/app/config/glance.yml:ro" \
	"glanceapp/glance:latest"
