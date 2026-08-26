# Backend deployment on Raspberry Pi OS

This deployment runs the Go backend as a `systemd` service that starts on boot
and restarts automatically after a crash.

## Prerequisites

- Raspberry Pi OS with `systemd`.
- Either Go installed on the Pi, or a Linux binary cross-compiled elsewhere.
- The frontend deployment for the kiosk page. See `deploy/frontend/README.md`.

The backend has no CGO dependencies, so it cross-compiles cleanly from any
machine. From the repository root on a development machine:

```sh
make build-pi
```

That writes `backend/kiosk` for `linux/arm64`. Copy the repository (or at least
`backend/kiosk` and `deploy/backend`) to the Pi.

## Install

On the Pi, from the repository root:

```sh
./deploy/backend/install.sh
```

The installer builds the binary with the local Go toolchain. To install a
binary that was cross-compiled elsewhere, pass its path instead:

```sh
./deploy/backend/install.sh ./backend/kiosk
```

The installer:

- installs the binary to `/usr/local/bin/kiosk-home-display`
- creates `/etc/kiosk-home-display/kiosk.env` from the template, without
  overwriting an existing file
- creates `/var/lib/kiosk-home-display` as the service working directory
- installs and enables `kiosk-home-display.service`

The service runs as the user that invoked the installer. Override it with
`KIOSK_USER=someone ./deploy/backend/install.sh`.

## Configuration

Edit `/etc/kiosk-home-display/kiosk.env` and restart the service:

```sh
sudo nano /etc/kiosk-home-display/kiosk.env
sudo systemctl restart kiosk-home-display
```

Keep `HTTP_ALLOWED_ORIGINS=localhost` so the WebSocket endpoint accepts the
kiosk page that Nginx serves on port `80`.

`DISPLAY_MODE=linux` drives the real display through `xset`. It only works when
the service user owns the X session, and it needs `DISPLAY` and `XAUTHORITY`
set in the env file. Install the backend as the same auto-login user that the
frontend deployment uses. `DISPLAY_BRIGHTNESS` is applied with `xrandr`, as a
software gamma adjustment on every connected output, so it works on HDMI
screens that expose no backlight device. Both tools come from the
`x11-xserver-utils` package. When `xrandr` is missing or no output is
connected, the service logs that brightness is unsupported and starts anyway.
Keep `DISPLAY_MODE=virtual` to run without touching the display hardware.

The SQLite database lives at `/var/lib/kiosk-home-display/data/kiosk.db`,
because the backend resolves it relative to the working directory.

## Update

```sh
git pull
./deploy/backend/install.sh
```

The installer rebuilds, replaces the binary, and restarts the service. The env
file and the database survive the update.

## Operate

```sh
systemctl status kiosk-home-display
journalctl -u kiosk-home-display -f
sudo systemctl restart kiosk-home-display
```
