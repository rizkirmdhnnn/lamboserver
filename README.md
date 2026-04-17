# LamboServer

Open-source local development environment manager for macOS. Manage PHP, Node.js, Nginx, DNS, and SSL — all from a single native desktop app.

Built with [Wails](https://wails.io) (Go + React).

![macOS](https://img.shields.io/badge/platform-macOS-lightgrey)
![Go](https://img.shields.io/badge/Go-1.23-00ADD8)
![React](https://img.shields.io/badge/React-18-61DAFB)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

### PHP Management
- Install and manage multiple PHP versions (7.4 – 8.5)
- Switch active PHP version with one click
- Automatic shell integration (`php`, `composer` available in terminal)
- PHP-FPM lifecycle control (start / stop / restart)
- Detects existing system PHP installations

### Node.js Management
- Install and manage multiple Node.js versions
- Switch active version instantly
- Downloads official releases from nodejs.org
- Shell symlinks for `node` and `npm`

### Site Management
- Link project folders as local `.test` domains
- Automatic Nginx server block generation
- Choose custom document root (e.g. `/public` for Laravel)
- SSL support with automatic certificate generation

### Web Server (Nginx)
- Embedded Nginx binary — no Homebrew required
- Runs as a macOS LaunchDaemon (root) for port 80 binding
- Auto-start on boot with KeepAlive
- Hot-reload configuration without restart

### DNS (dnsmasq)
- Embedded dnsmasq binary
- All `*.test` domains resolve to `127.0.0.1`
- Automatic `/etc/resolver/test` setup
- Runs as a macOS LaunchDaemon for port 53

### SSL Certificates
- Local Certificate Authority (CA) generation
- Auto-trust CA in macOS Keychain
- Per-site SSL certificate generation
- 10-year certificate validity

### Developer Experience
- Dashboard with service health overview
- Log viewer for Nginx, PHP-FPM, and dnsmasq
- Debug mode with detailed action logging
- One-time admin password prompt (sudoers-based privilege escalation)
- Shell integration install/uninstall

## Screenshots

| Dashboard | Sites | PHP |
|-----------|-------|-----|
| *coming soon* | *coming soon* | *coming soon* |

## Requirements

- **macOS** (Apple Silicon or Intel)
- **Go** 1.23+
- **Node.js** (for frontend development only)
- **Wails CLI** v2

## Quick Start

### Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Clone & Build

```bash
git clone https://github.com/rizkirmdhnnn/lamboserver.git
cd lamboserver
wails build
```

The built app will be at `build/bin/LamboServer.app`.

### Development Mode

```bash
wails dev
```

This starts a Vite dev server with hot-reload for the frontend and live Go backend recompilation.

## Architecture

```
lamboserver/
├── main.go                     # Wails app entry point
├── app.go                      # App struct, startup/shutdown, service orchestration
├── app_*.go                    # Feature-specific API methods (bound to frontend)
├── internal/
│   ├── system/
│   │   ├── launchd.go          # macOS LaunchDaemon/Agent management
│   │   ├── helper.go           # Privileged helper (sudoers-based)
│   │   ├── paths.go            # Centralized file path management
│   │   ├── binary.go           # Binary locator and embedded extraction
│   │   ├── darwin.go           # macOS-specific utilities (admin prompts, keychain)
│   │   ├── shell.go            # Shell RC integration (~/.zshrc, ~/.bashrc)
│   │   ├── download.go         # HTTP download with progress
│   │   └── embedded/           # Bundled nginx and dnsmasq binaries
│   ├── nginx/                  # Nginx config generation and lifecycle
│   ├── dns/                    # dnsmasq config and DNS resolver setup
│   ├── php/                    # PHP version management, FPM, detection
│   ├── node/                   # Node.js version management
│   ├── site/                   # Site linking, Nginx vhost generation
│   ├── cert/                   # CA and SSL certificate management
│   ├── config/                 # Thread-safe JSON config store
│   ├── debug/                  # Debug logger
│   └── log/                    # Log file reader
└── frontend/
    └── src/
        ├── App.tsx             # Tab-based navigation
        └── pages/              # Dashboard, Sites, Services, PHP, Node, Logs
```

### How It Works

1. **First Launch**: Installs a privileged helper script and sudoers entry (one-time admin password prompt)
2. **Service Management**: Nginx and dnsmasq run as macOS LaunchDaemons (root) for privileged port binding. PHP-FPM runs as a user-level LaunchAgent
3. **Version Switching**: PHP and Node.js versions are managed via symlinks in `~/.lamboserver/bin/`
4. **Site Linking**: Creates Nginx server blocks pointing to your project directory with optional SSL
5. **Configuration**: All state persisted in `~/.lamboserver/config.json`

### Data Directory

All runtime data lives in `~/.lamboserver/`:

```
~/.lamboserver/
├── bin/              # Symlinks and helper scripts
├── nginx/            # Nginx config, sites, logs
├── dnsmasq/          # dnsmasq config
├── php/              # Installed PHP versions
├── node/             # Installed Node.js versions
├── certs/            # CA and site certificates
├── logs/             # Service logs
└── config.json       # App configuration
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Desktop Framework | Wails v2 |
| Backend | Go 1.23 |
| Frontend | React 18 + TypeScript |
| Bundler | Vite |
| Icons | Lucide React |
| Process Management | macOS launchd |
| Privilege Escalation | sudoers + helper script |

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License. See [LICENSE](LICENSE) for details.

## Author

**Achmad Rizki Ramadhan** — [achmadrizkiramadhan0101@gmail.com](mailto:achmadrizkiramadhan0101@gmail.com)
