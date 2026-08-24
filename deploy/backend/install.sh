#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
binary=${1:-}
service_user=${KIOSK_USER:-${SUDO_USER:-$(id -un)}}
service_name=kiosk-home-display
config_dir=/etc/kiosk-home-display
data_dir=/var/lib/kiosk-home-display
env_file="$config_dir/kiosk.env"

if ! id "$service_user" >/dev/null 2>&1; then
  echo "User $service_user does not exist. Set KIOSK_USER to an existing user." >&2
  exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
  echo "Systemd is required to run the backend as a service." >&2
  exit 1
fi

if [ -z "$binary" ]; then
  if ! command -v go >/dev/null 2>&1; then
    echo "Go is required to build the backend. Install it with: sudo apt install golang" >&2
    echo "Or cross-compile elsewhere and pass the binary: $0 /path/to/kiosk" >&2
    exit 1
  fi

  echo "Building the backend..."
  build_dir=$(mktemp -d)
  trap 'rm -rf "$build_dir"' EXIT INT TERM
  binary="$build_dir/kiosk"
  (cd "$repo_root/backend" && CGO_ENABLED=0 go build -o "$binary" ./cmd/kiosk)
fi

if [ ! -f "$binary" ]; then
  echo "Backend binary not found at $binary." >&2
  exit 1
fi

service_group=$(id -gn "$service_user")

sudo install -d -m 755 "$config_dir"
sudo install -d -m 755 -o "$service_user" -g "$service_group" "$data_dir"
sudo install -d -m 755 -o "$service_user" -g "$service_group" "$data_dir/data"

if [ -f "$env_file" ]; then
  echo "Keeping the existing $env_file."
else
  sudo install -m 640 -o root -g "$service_group" \
    "$script_dir/kiosk.env.example" \
    "$env_file"
  echo "Created $env_file from the template. Review it before going live."
fi

sudo install -m 755 "$binary" "/usr/local/bin/$service_name"

sed "s|__SERVICE_USER__|$service_user|g" \
  "$script_dir/systemd/$service_name.service" \
  | sudo tee "/etc/systemd/system/$service_name.service" >/dev/null
sudo chmod 644 "/etc/systemd/system/$service_name.service"

sudo systemctl daemon-reload
sudo systemctl enable "$service_name"
sudo systemctl restart "$service_name"

echo "Backend installed and running as $service_user."
echo "Check it with: systemctl status $service_name"
