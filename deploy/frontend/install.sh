#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
dist_dir=${1:-"$repo_root/frontend/dist"}
web_root=/var/www/kiosk-home-display

if [ ! -f "$dist_dir/index.html" ]; then
  echo "Frontend build not found at $dist_dir. Run 'pnpm build' in frontend first." >&2
  exit 1
fi

if ! command -v nginx >/dev/null 2>&1; then
  echo "Nginx is required. Install it with: sudo apt install nginx" >&2
  exit 1
fi

if ! command -v chromium >/dev/null 2>&1 && ! command -v chromium-browser >/dev/null 2>&1; then
  echo "Chromium is required. Install it with: sudo apt install chromium" >&2
  exit 1
fi

sudo install -d -m 755 "$web_root"
sudo cp -R "$dist_dir/." "$web_root/"
sudo install -m 644 \
  "$script_dir/nginx/kiosk-home-display.conf" \
  /etc/nginx/sites-available/kiosk-home-display
sudo ln -sf \
  /etc/nginx/sites-available/kiosk-home-display \
  /etc/nginx/sites-enabled/kiosk-home-display
sudo rm -f /etc/nginx/sites-enabled/default
sudo install -m 755 \
  "$script_dir/scripts/kiosk-home-display-browser" \
  /usr/local/bin/kiosk-home-display-browser
sudo install -m 644 \
  "$script_dir/autostart/kiosk-home-display.desktop" \
  /etc/xdg/autostart/kiosk-home-display.desktop
sudo nginx -t
sudo systemctl reload nginx

echo "Frontend installed. Log in to the Raspberry Pi desktop session to start Chromium in kiosk mode."
