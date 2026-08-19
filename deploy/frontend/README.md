# Frontend deployment on Raspberry Pi OS

This deployment serves the Vite build with Nginx and starts Chromium in kiosk
mode when a user logs in to the Raspberry Pi desktop session.

## Prerequisites

- Raspberry Pi OS with Desktop and an automatic-login user.
- `nginx` and `chromium` installed. On some Raspberry Pi OS releases the
  Chromium command is named `chromium-browser`; the launcher supports both.
- The backend running locally on port `8080`.

Configure the backend with `HTTP_ALLOWED_ORIGINS=localhost` so its WebSocket
endpoint accepts requests from the kiosk page served by Nginx.

Install the operating-system dependencies once:

```sh
sudo apt update
sudo apt install -y nginx chromium
```

If the Chromium package is named differently on the installed Raspberry Pi OS
release, install `chromium-browser` instead.

## Install

From the repository root:

```sh
cd frontend
pnpm install --frozen-lockfile
pnpm build

cd ..
./deploy/frontend/install.sh
```

The installer copies `frontend/dist` to `/var/www/kiosk-home-display`, enables
the Nginx site on port `80`, and installs a desktop autostart entry. It removes
Nginx's default enabled site, so this Raspberry Pi should be dedicated to the
kiosk display.

On the next desktop login Chromium opens `http://127.0.0.1/` in kiosk mode.
If Chromium crashes, the launcher retries after two seconds.

## Update

Build the frontend again, then re-run the installer:

```sh
cd frontend && pnpm build
cd .. && ./deploy/frontend/install.sh
```
