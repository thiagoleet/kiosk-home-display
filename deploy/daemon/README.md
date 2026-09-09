# Daemon deployment on Raspberry Pi OS

This deployment runs the Go daemon as a `systemd` service that starts on boot
and restarts automatically after a crash.

## Prerequisites

- Raspberry Pi OS with `systemd`.
- Either Go installed on the Pi, or a Linux binary cross-compiled elsewhere.
- The frontend deployment for the kiosk page. See `deploy/frontend/README.md`.

The daemon has no CGO dependencies, so it cross-compiles cleanly from any
machine. From the repository root on a development machine:

```sh
make build-pi
```

That writes `daemon/kiosk` for `linux/arm64`. Copy the repository (or at least
`daemon/kiosk` and `deploy/daemon`) to the Pi.

## Install

On the Pi, from the repository root:

```sh
./deploy/daemon/install.sh
```

The installer builds the binary with the local Go toolchain. To install a
binary that was cross-compiled elsewhere, pass its path instead:

```sh
./deploy/daemon/install.sh ./daemon/kiosk
```

The installer:

- installs the binary to `/usr/local/bin/kiosk-home-display`
- creates `/etc/kiosk-home-display/kiosk.env` from the template, without
  overwriting an existing file
- creates `/var/lib/kiosk-home-display` as the service working directory
- installs and enables `kiosk-home-display.service`

The service runs as the user that invoked the installer. Override it with
`KIOSK_USER=someone ./deploy/daemon/install.sh`.

## Configuration

Edit `/etc/kiosk-home-display/kiosk.env` and restart the service:

```sh
sudo nano /etc/kiosk-home-display/kiosk.env
sudo systemctl restart kiosk-home-display
```

Keep `HTTP_ALLOWED_ORIGINS=localhost` so the WebSocket endpoint accepts the
kiosk page that Nginx serves on port `80`.

### Display mode

`DISPLAY_MODE` decides how the screen is powered, and it has to match the
session the Pi actually runs. Check it first:

```sh
loginctl list-sessions
loginctl show-session <id-on-seat0> -p Type -p Name
```

Read the session on `seat0`, not the one you are typing in: over SSH
`$XDG_SESSION_TYPE` reports `tty` no matter what the console runs.

- `x11` → `DISPLAY_MODE=linux`. Powers the screen with `xset dpms`, and applies
  `DISPLAY_BRIGHTNESS` with `xrandr` as a software gamma adjustment on every
  connected output, so it works on HDMI screens that expose no backlight
  device. Both tools come from `x11-xserver-utils`. Needs `DISPLAY` and
  `XAUTHORITY` in the env file.
- `wayland` → `DISPLAY_MODE=wayland`. The default on Raspberry Pi OS Bookworm
  and later (labwc, wayfire), where `xset` cannot reach the physical output.
  Install `wlopm` for this mode:

  ```sh
  sudo apt install wlopm
  ```

  `wlopm` speaks `zwlr_output_power_management_v1`, the Wayland equivalent of
  DPMS: the output keeps its mode and position and only the sink powers down.
  It is effectively required. Without it, sleeping fails with a message saying
  so, because the only alternative — `wlr-randr` disabling the output — reflows
  every surface in the session and, on some compositor and display
  combinations, cannot be undone: re-enabling answers `failed to apply
  configuration` and the screen stays dark until the session restarts. Since
  the daemon sleeps the display on a timer, that would repeat every
  `IDLE_TIMEOUT`. Set `WAYLAND_ALLOW_OUTPUT_DISABLE=true` to use the fallback
  anyway, on a host where it is known to work.

  `wlopm` is packaged from Debian trixie on (Raspberry Pi OS 13). On Bookworm
  it has to be built from source, and an X11 session with `DISPLAY_MODE=linux`
  is the easier path there.

  Waking re-enables any output that reports `Enabled: no` before powering the
  sinks back on, so a screen the fallback switched off recovers on the next
  wake instead of needing the session restarted.

  No env file entries are required: a system service inherits no session
  variables, so the daemon derives them from the user it runs as —
  `/run/user/<uid>` for the runtime directory and whichever `wayland-*` socket
  sits inside it. Set `XDG_RUNTIME_DIR` and `WAYLAND_DISPLAY` only to override
  that. There is no brightness control on this path, so the service logs that
  brightness is unsupported and starts anyway.
- `virtual` → logs the transitions and touches no hardware. The API still
  answers `200`, which makes this the quietest way for a box to look healthy
  while the screen never turns off.

Either real mode needs the service user to own the desktop session, so install
the daemon as the same auto-login user the frontend deployment uses.

The installer never overwrites an existing `/etc/kiosk-home-display/kiosk.env`,
so a box installed before this setting existed keeps `virtual` across every
update. It now prints the mode it left in place — check that line after
deploying.

On X11 the daemon enables DPMS before powering the screen down, and zeroes the
three DPMS timeouts while doing it. A kiosk session usually disables DPMS to
stop the screen blanking on its own (`raspi-config` writes `X -s 0 -dpms` for
that), and in that state `xset dpms force off` exits `0` and leaves the screen
lit. Zeroed timeouts keep the automatic blanking away, so only this daemon
powers the screen down.

### Verifying the display actually powers off

```sh
curl -s -X POST localhost:8080/api/display/sleep
journalctl -u kiosk-home-display -n 20
```

`{"status":"ok"}` now means the tools ran and reported success: both transitions
drive the controller even when the daemon already believes the screen is in that
state. That belief resets to on at every restart, so trusting it used to answer
`200` on the first wake after a deploy without running anything.

A `[DISPLAY] sleep` line means the service is in `virtual` mode. Otherwise the
error names what is missing: the package, the X session, or the compositor
socket. `the compositor refused the output configuration` means the `wlr-randr`
fallback is in use and the compositor will not take it — install `wlopm`.
`no wayland socket in /run/user/<uid>` means the compositor is not running as
the service user — either the desktop session belongs to another
user, or the Pi boots to a console with no compositor at all, in which case
neither real mode can power the screen off. Confirm with:

```sh
ls /run/user/$(id -u)/wayland-*
```

### Schedule and idle timeout

The scheduler applies the window the current time already sits in when the
service starts, so restarting during the off window — which every deploy after
`SCHEDULE_OFF` is — turns the display off right away instead of waiting for the
next `SCHEDULE_ON`. It also settles the display on any later tick, so a boundary
minute missed by a busy process, or skipped by the clock jump a Pi without an
RTC makes when NTP lands, no longer strands the screen in the wrong state.

The idle timeout keeps watching after it fires, and every wake restarts its
countdown. Nothing reports user activity yet, so with `IDLE_ENABLED=true` the
display stays on for `IDLE_TIMEOUT` after each wake and then sleeps again —
including the wake at `SCHEDULE_ON`. Set `IDLE_ENABLED=false` to keep the screen
on for the whole on window.

The SQLite database lives at `/var/lib/kiosk-home-display/data/kiosk.db`,
because the daemon resolves it relative to the working directory.

## Update

```sh
git pull
./deploy/daemon/install.sh
```

The installer rebuilds, replaces the binary, and restarts the service. The env
file and the database survive the update.

## Operate

```sh
systemctl status kiosk-home-display
journalctl -u kiosk-home-display -f
sudo systemctl restart kiosk-home-display
```
