![Thiago - Capa](https://user-images.githubusercontent.com/9437391/153274659-915c4df9-0032-4757-a9a2-6a85107c276b.png)

# Hello there!

## Who Am I?

- 🇧🇷 I'm from Brazil
- 👨‍💻Software Engineer, currently working with Frontend Development
- 💡 Always learning.
- ⚙️ Contact me on [LinkedIn](https://www.linkedin.com/in/thiagofmleite/)
- 🚶‍♂️Follow me on [Instagram](https://instagram.com/thiagoleet) and [Twitch](https://twitch.tv/thiagoleet).

# Kiosk Home Display

**Kiosk Home Display** is a Linux-based home display system designed to run continuously on a dedicated screen, such as a Raspberry Pi connected to a TV or monitor.

The system combines a **Go daemon** with a **React frontend** to provide an event-driven display that can show an animated idle screen, notifications, system information, and real-time events.

## Key Features

- 🖥️ Fullscreen kiosk interface
- ✨ Animated idle/screensaver display
- ⏱️ Configurable idle timeout
- 🗓️ Scheduled display on/off times
- 🔔 Event-driven notifications
- 🖨️ Printer/CUPS event monitoring
- 📡 Extensible event sources such as MQTT
- ⚡ Real-time communication between daemon and frontend
- 💡 Display power and sleep control
- 🔄 Automatic startup and recovery through `systemd`
- 🐧 Designed for Linux and Raspberry Pi environments

## Architecture

The project is organized as a monorepo containing two main applications:

```text
|kiosk-home-display/
├── daemon/ # Go daemon and system integrations
├── frontend/ # React-based kiosk interface
├── deploy/ # Linux/systemd deployment configuration
└── docs/ # Architecture and project documentation
```

The Go daemon acts as the system's event and display controller. It monitors external events, manages display state and scheduling, and communicates with the React frontend through WebSockets.

```text
             ┌─────────────────────┐
             │      Linux OS       │
             └──────────┬──────────┘
                        │
        ┌───────────────┼────────────────┐
        │               │                │
       CUPS            MQTT           Scheduler
        │               │                │
        └───────────────┼────────────────┘
                        ▼
               ┌────────────────┐
               │   Go Daemon    │
               │                │
               │  Event Bus     │
               │  Display Mgmt  │
               │  WebSocket     │
               └───────┬────────┘
                       │
                    WebSocket
                       │
                       ▼
               ┌────────────────┐
               │ React Frontend │
               │                │
               │ Idle Screen    │
               │ Notifications  │
               │ Dashboard      │
               └────────────────┘
```

The project is intended to be **modular and extensible**, allowing new event sources, display behaviors, and UI screens to be added without tightly coupling the system components.
