# Deploy Pi-hole Control on Raspberry Pi OS

The Pi needs systemd, Node.js 20.9 or newer, pnpm, and package registry access during the build.

| Service | Port |
| --- | --- |
| Pi-hole Control | 80 |
| Kiosk frontend (Nginx) | 8082 |
| Kiosk Go daemon | 8080 |
| Pi-hole web interface and API | 8181 |

On an existing kiosk, deploy the frontend first to free port 80. Use the same `KIOSK_ENV` you already deploy with, then log out and back in to reopen Chromium at its new address.

```sh
make deploy-frontend KIOSK_ENV=snespi
make deploy-pihole-control
sudoedit /etc/pihole-control/pihole-control.env
sudo systemctl restart pihole-control
```

Set `PIHOLE_URL=http://127.0.0.1:8181` and set `PIHOLE_PASSWORD` to your Pi-hole application password. Leave it blank if Pi-hole has no password. The installer creates a root-only environment file and preserves it on updates, so check the URL if upgrading an existing installation.

Open <http://milkpi.local/> from your local network or <http://127.0.0.1/> on the Pi. Pi-hole Control does not prompt for a login, so any device that can reach the service can pause blocking once the application password is configured.

Check the service and read the blocking status on the Pi:

```sh
systemctl status pihole-control
journalctl -u pihole-control -n 30
curl --fail http://127.0.0.1/api/blocking
```

Re-run `make deploy-pihole-control` to update. The installer builds before switching the active release and restarting the service. Releases are kept under `/opt/pihole-control/releases`. The active release is the target of `/opt/pihole-control/current`.
