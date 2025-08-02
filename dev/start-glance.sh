#!/usr/bin/env sh

set -e

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
PODMAN=$(which "podman" 2> /dev/null || which "docker")

set -x

${PODMAN} run \
	--pull=always \
	--rm \
	--network=host \
	--volume "${SCRIPT_DIR}/glance.yml:/app/config/glance.yml:ro" \
	"glanceapp/glance:latest"
