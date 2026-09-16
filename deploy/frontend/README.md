# Frontend deployment on Raspberry Pi OS

This deployment serves the Vite build with Nginx and starts Chromium in kiosk
mode when a user logs in to the Raspberry Pi desktop session.

## Prerequisites

- Raspberry Pi OS with Desktop and an automatic-login user.
- `nginx` and `chromium` installed. On some Raspberry Pi OS releases the
  Chromium command is named `chromium-browser`; the launcher supports both.
- The daemon running locally on port `8080`.

Configure the daemon with `HTTP_ALLOWED_ORIGINS=localhost` so its WebSocket
endpoint accepts requests from the kiosk page served by Nginx.

Install the operating-system dependencies once:

```sh
sudo apt update
sudo apt install -y nginx chromium
```

If the Chromium package is named differently on the installed Raspberry Pi OS
release, install `chromium-browser` instead.

## Install

From the repository root, on the box being deployed:

```sh
cd frontend && pnpm install --frozen-lockfile && cd ..
make deploy-frontend KIOSK_ENV=snespi
```

That builds with `deploy/environments/snespi/frontend.env` and installs the
result. The profile it names decides the theme, the locale and which features
the screen offers, and Vite inlines it at build time — so the artifact belongs
to that box and cannot be copied to the other one.

Without an environment the build falls back to the `milkpi` profile:

```sh
cd frontend && pnpm build
cd .. && ./deploy/frontend/install.sh
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
make deploy-frontend KIOSK_ENV=snespi
```

Check the browser console after adding or renaming an environment: an unknown
`VITE_KIOSK_PROFILE` does not fail the build, it falls back to `milkpi` and
logs the unknown id.
