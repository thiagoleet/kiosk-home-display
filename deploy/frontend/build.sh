#!/bin/sh

# Builds the frontend for one environment. The profile and every other VITE_
# setting are inlined by Vite at build time, so the artifact this produces
# belongs to one box and cannot be reused on the other.

set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
environment=${1:-${KIOSK_ENV:-}}

# Only the directories are environments; the folder also holds a README.
list_environments() {
  find "$repo_root/deploy/environments" -mindepth 1 -maxdepth 1 -type d \
    -exec basename {} \; | sort | sed 's/^/  /'
}

if [ -z "$environment" ]; then
  echo "Usage: $0 <environment>" >&2
  echo "Or set KIOSK_ENV. Available:" >&2
  list_environments >&2
  exit 1
fi

env_file="$repo_root/deploy/environments/$environment/frontend.env"

if [ ! -f "$env_file" ]; then
  echo "No frontend.env for environment $environment at $env_file." >&2
  echo "Available:" >&2
  list_environments >&2
  exit 1
fi

# set -a exports every assignment the file makes, which is how the VITE_
# settings reach the Vite process.
set -a
. "$env_file"
set +a

echo "Building the frontend for $environment (VITE_KIOSK_PROFILE=${VITE_KIOSK_PROFILE:-unset})..."

cd "$repo_root/frontend"
pnpm build
