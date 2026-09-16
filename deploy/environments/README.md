# Environments

Each Raspberry Pi running this kiosk is an *environment*: a folder here holding
the settings that belong to that box and nothing else.

```
deploy/environments/
  snespi/
    frontend.env   # VITE_ settings, inlined into the build
    kiosk.env      # daemon settings, the template for /etc/kiosk-home-display/kiosk.env
  milkpi/
    frontend.env
    kiosk.env
```

## What separates them today

| | snespi | milkpi |
| --- | --- | --- |
| Printer | none | attached |
| `PRINTER_MODE` | `off` | `cups` |
| `VITE_KIOSK_PROFILE` | `snespi` | `milkpi` |
| `features.printer` | `false` | `true` |

Both halves have to agree. The daemon side decides whether print events are
produced at all; the frontend profile decides whether the screen has anywhere
to show them. A box with `PRINTER_MODE=off` and a profile that still claims a
printer would sit on "ready" for good, describing a device that is not there.

## Deploying one box

From the repository root, on the box itself:

```sh
make deploy-snespi
make deploy-milkpi
```

That builds the frontend with that environment's `frontend.env`, installs it,
then installs the daemon with that environment's `kiosk.env`. The halves can
also be run on their own:

```sh
make deploy-frontend KIOSK_ENV=snespi
make deploy-daemon KIOSK_ENV=snespi
```

Running a target with no `KIOSK_ENV` keeps the older behaviour: the frontend
builds with its defaults (which fall back to the `milkpi` profile) and the
daemon installs from `deploy/daemon/kiosk.env.example`.

## The frontend build belongs to one box

Vite inlines every `VITE_` setting at build time, so `frontend/dist` after
`make deploy-snespi` is a snespi artifact. It cannot be copied to the other
box, and changing a value in `frontend.env` means rebuilding.

An unknown `VITE_KIOSK_PROFILE` does not fail the build — it falls back to
`milkpi` and logs the unknown id to the browser console, so check the console
after adding an environment.

## The daemon env file belongs to the box, not to git

`kiosk.env` here is a *template*. The installer copies it to
`/etc/kiosk-home-display/kiosk.env` on a first install and never touches it
again, because that file carries whatever was tuned on the box by hand —
`DISPLAY_MODE`, the Wayland socket, weather coordinates. Editing the template
in git therefore changes nothing on a box that is already installed; the
installer says so when it skips one.

To push the template over the installed file:

```sh
KIOSK_ENV_OVERWRITE=true make deploy-daemon KIOSK_ENV=snespi
```

The previous file is kept at `/etc/kiosk-home-display/kiosk.env.bak`.

Each install prints the `DISPLAY_MODE` and `PRINTER_MODE` in effect, which is
the quickest way to catch a box still running an older file.

## Adding an environment

1. Copy a folder here and rename it.
2. Set `VITE_KIOSK_PROFILE` in `frontend.env` and `PRINTER_MODE` in
   `kiosk.env`.
3. Add the matching profile in `frontend/src/profiles/` and register it in
   `frontend/src/profiles/index.ts` — the id there must match
   `VITE_KIOSK_PROFILE`.
4. Add a `deploy-<name>` target to the `Makefile` if you want the shorthand.
