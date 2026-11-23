#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
app_dir="$script_dir/../../pihole-control"
install_dir=/opt/pihole-control
config_dir=/etc/pihole-control
env_file="$config_dir/pihole-control.env"
service=pihole-control.service

if [ "$(uname -s)" != Linux ] || ! command -v systemctl >/dev/null 2>&1; then
  echo "Run this installer on a Linux host with systemd." >&2
  exit 1
fi

if ! command -v node >/dev/null 2>&1 || ! command -v pnpm >/dev/null 2>&1; then
  echo "Install Node.js and pnpm before deploying Pi-hole Control." >&2
  exit 1
fi
if [ "$(id -u)" -eq 0 ]; then
  echo "Run this installer as a regular user with sudo access." >&2
  exit 1
fi

node_path=$(node -p 'process.execPath')
build_dir=$(mktemp -d)
trap 'rm -rf -- "$build_dir"' EXIT
trap 'exit 1' HUP INT TERM

# Build with only tracked inputs. Pi-hole credentials are injected at runtime.
cp -R "$app_dir/app" "$build_dir/"
for file in package.json pnpm-lock.yaml tsconfig.json next.config.mjs postcss.config.mjs; do
  cp "$app_dir/$file" "$build_dir/"
done
(
  cd "$build_dir"
  NODE_ENV=development pnpm install --frozen-lockfile
  pnpm build
)
artifact="$build_dir/.next/standalone"
test -f "$artifact/server.js"
cp -R "$build_dir/.next/static" "$artifact/.next/static"

sudo install -d -m 0755 "$install_dir/releases"
release=$(sudo mktemp -d "$install_dir/releases/release.XXXXXXXX")
sudo cp -R "$artifact/." "$release/"
sudo chmod 0755 "$release"
# Copy the build's Node executable so systemd can use it with version managers
# that install Node inside the invoking user's home directory.
sudo install -m 0755 "$node_path" "$release/node"

sudo install -d -m 0700 "$config_dir"
if ! sudo test -e "$env_file"; then
  sudo install -m 0600 "$script_dir/pihole-control.env.example" "$env_file"
  echo "Created $env_file. Set the Pi-hole URL and application password there."
fi

sudo ln -s "$release" "$release/current-link"
sudo mv -Tf "$release/current-link" "$install_dir/current"
sudo install -m 0644 "$script_dir/systemd/$service" "/etc/systemd/system/$service"
sudo systemctl daemon-reload
sudo systemctl enable "$service"
sudo systemctl restart "$service"

echo "Pi-hole Control installed at http://127.0.0.1/"
echo "Check it with: systemctl status $service"
