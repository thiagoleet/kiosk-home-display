#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
binary=${1:-}
environment=${KIOSK_ENV:-}
overwrite_env=${KIOSK_ENV_OVERWRITE:-false}
service_user=${KIOSK_USER:-${SUDO_USER:-$(id -un)}}
service_name=kiosk-home-display
config_dir=/etc/kiosk-home-display
data_dir=/var/lib/kiosk-home-display
env_file="$config_dir/kiosk.env"

# Only the directories are environments; the folder also holds a README.
list_environments() {
  find "$repo_root/deploy/environments" -mindepth 1 -maxdepth 1 -type d \
    -exec basename {} \; | sort | sed 's/^/  /'
}

# Each box has its own template under deploy/environments. Without KIOSK_ENV
# the generic example is used, which is what an install did before the
# environments existed.
if [ -n "$environment" ]; then
  env_template="$repo_root/deploy/environments/$environment/kiosk.env"

  if [ ! -f "$env_template" ]; then
    echo "No environment named $environment at $env_template." >&2
    echo "Available:" >&2
    list_environments >&2
    exit 1
  fi
else
  env_template="$script_dir/kiosk.env.example"
fi

if ! id "$service_user" >/dev/null 2>&1; then
  echo "User $service_user does not exist. Set KIOSK_USER to an existing user." >&2
  exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
  echo "Systemd is required to run the daemon as a service." >&2
  exit 1
fi

if [ -z "$binary" ]; then
  if ! command -v go >/dev/null 2>&1; then
    echo "Go is required to build the daemon. Install it with: sudo apt install golang" >&2
    echo "Or cross-compile elsewhere and pass the binary: $0 /path/to/kiosk" >&2
    exit 1
  fi

  echo "Building the daemon..."
  build_dir=$(mktemp -d)
  trap 'rm -rf "$build_dir"' EXIT INT TERM
  binary="$build_dir/kiosk"
  (cd "$repo_root/daemon" && CGO_ENABLED=0 go build -o "$binary" ./cmd/kiosk)
fi

if [ ! -f "$binary" ]; then
  echo "Daemon binary not found at $binary." >&2
  exit 1
fi

service_group=$(id -gn "$service_user")

sudo install -d -m 755 "$config_dir"
sudo install -d -m 755 -o "$service_user" -g "$service_group" "$data_dir"
sudo install -d -m 755 -o "$service_user" -g "$service_group" "$data_dir/data"

if [ -f "$env_file" ] && [ "$overwrite_env" = "true" ]; then
  # Opt-in only: the file on the box is the one that carries whatever was
  # tuned there by hand, and a template cannot know about it.
  sudo cp "$env_file" "$env_file.bak"
  sudo install -m 640 -o root -g "$service_group" \
    "$env_template" \
    "$env_file"
  echo "Replaced $env_file from $env_template."
  echo "  The previous file is at $env_file.bak."
elif [ -f "$env_file" ]; then
  echo "Keeping the existing $env_file."
  # An env file from an earlier install survives every update, so a box can sit
  # on DISPLAY_MODE=virtual for good: the API answers 200 and logs the
  # transition while the screen never powers down. Report what is in effect.
  installed_mode=$(sudo sed -n 's/^DISPLAY_MODE=//p' "$env_file" | tail -n 1)
  installed_mode=${installed_mode:-virtual (unset)}
  echo "  DISPLAY_MODE=$installed_mode"
  case "$installed_mode" in
    linux | wayland) ;;
    *)
      echo "  Warning: this mode never touches the display hardware." >&2
      echo "  Set DISPLAY_MODE=linux (X11) or wayland (labwc/wayfire) to power the screen off." >&2
      ;;
  esac
  # The printer mode is what separates the boxes, so it is worth the same
  # check: a box that kept an older file can be on the wrong one.
  installed_printer=$(sudo sed -n 's/^PRINTER_MODE=//p' "$env_file" | tail -n 1)
  echo "  PRINTER_MODE=${installed_printer:-virtual (unset)}"

  if [ -n "$environment" ]; then
    # The installed file wins, so an environment file edited in git changes
    # nothing on a box that already has one. Say so rather than let the deploy
    # look like it applied.
    echo "  $env_template was not applied. Pass KIOSK_ENV_OVERWRITE=true to replace it."
  fi
else
  sudo install -m 640 -o root -g "$service_group" \
    "$env_template" \
    "$env_file"
  echo "Created $env_file from $env_template. Review it before going live."
fi

sudo install -m 755 "$binary" "/usr/local/bin/$service_name"

sed "s|__SERVICE_USER__|$service_user|g" \
  "$script_dir/systemd/$service_name.service" \
  | sudo tee "/etc/systemd/system/$service_name.service" >/dev/null
sudo chmod 644 "/etc/systemd/system/$service_name.service"

sudo systemctl daemon-reload
sudo systemctl enable "$service_name"
sudo systemctl restart "$service_name"

echo "Daemon installed and running as $service_user."
echo "Check it with: systemctl status $service_name"
