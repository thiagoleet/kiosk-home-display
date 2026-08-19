# Kiosk Home Display

## Architecture

The project has two applications:

- backend: Go
- frontend: React + Vite + TypeScript

## Frontend

- React
- Vite
- TypeScript
- Lucide React
- WebSocket for real-time events
- No Ant Design
- No SCSS

## UI architecture

KioskLayout is the persistent shell.

KioskHeader is outside KioskTransition and must remain
static during screen transitions.

KioskTransition controls transitions between:

- home
- notification

HomeLayout and NotificationLayout are content layouts.

## Notifications

Notification has:

- id
- type
- level
- context
- title
- message
- duration

NotificationContext identifies the domain:

- printer
- system
- network
- display

NotificationLevel identifies semantic severity.

## Activities

Activities are independent from notifications.

A domain event may produce:

- notification
- activity

The frontend receives both through WebSocket.

## Themes

The project will eventually support different visual themes.

Current target:

- Modern Dark

Future:

- Retro 16-bit
