# Deploy Pi-hole Control on Raspberry Pi OS

Run this on the target Pi from the repository root as a user with sudo access. The Pi needs systemd, Node.js 20.9 or newer, pnpm, access to the package registry during the build, and network access to your Pi-hole v6 instance.

```sh
make deploy-pihole-control
sudoedit /etc/pihole-control/pihole-control.env
sudo systemctl restart pihole-control
```

Set `PIHOLE_URL` to the origin of the Pi-hole web interface and `PIHOLE_PASSWORD` to its application password. Leave the password blank if your Pi-hole has no password. The installer creates a root-only environment file and preserves it on updates.

Open <http://127.0.0.1:8081> on the Pi. To access it from another computer, forward the same local port over SSH and open the same URL in your browser:

```sh
ssh -N -L 127.0.0.1:8081:127.0.0.1:8081 user@your-pi
```

Check the service and read the blocking status:

```sh
systemctl status pihole-control
journalctl -u pihole-control -n 30
curl --fail http://127.0.0.1:8081/api/blocking
```

Re-run `make deploy-pihole-control` to update. The installer builds before switching the active release and restarts the service. Releases are kept under `/opt/pihole-control/releases`. The active release is the target of `/opt/pihole-control/current`.
