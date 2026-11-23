# Pi-hole Control

A standalone Next.js app for Pi-hole v6. The pxlkit interface shows the current blocking state and offers **Disable Block for 5 minutes**. Pi-hole owns the 300-second timer and resumes blocking even after you close the page.

## Run locally

Use Node.js 20.9 or newer.

```sh
npm ci
cp .env.example .env.local
```

Set `PIHOLE_URL` in `.env.local` to your Pi-hole web origin. Set `PIHOLE_PASSWORD` to a [Pi-hole application password](https://docs.pi-hole.net/api/auth/), or leave it blank if your Pi-hole has no password. Next.js reads the password only on the server.

```sh
npm run dev
```

Open <http://127.0.0.1:8080>. The app listens on loopback and the write route accepts same-origin requests from the local interface.
