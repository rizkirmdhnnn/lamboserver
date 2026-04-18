# LamboServer

Open-source local development environment manager for macOS. Manage PHP, Node.js, PostgreSQL, MySQL, Nginx, DNS, and SSL — all from a single native desktop app.

Built with [Wails](https://wails.io) (Go + React).

![macOS](https://img.shields.io/badge/platform-macOS-lightgrey)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8)
![React](https://img.shields.io/badge/React-18-61DAFB)
![License](https://img.shields.io/badge/license-MIT-green)

## Download

Grab the latest DMG from the [Releases](https://github.com/rizkirmdhnnn/lamboserver/releases) page.

> **Note:** LamboServer is ad-hoc signed (no Apple Developer certificate). On first launch, right-click the app and select "Open", then confirm in the dialog.

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

### Database Management
- **PostgreSQL** — install, start/stop, manage instances
- **pgweb** — built-in web-based PostgreSQL admin UI
- **MySQL** — install, start/stop, manage instances
- **phpMyAdmin** — built-in web-based MySQL admin UI

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

### System Tray
- Native macOS menu bar icon (monochrome template, Dark/Light mode aware)
- Live service status and start/stop/restart controls from the tray
- Quick access to recent `.test` sites — opens in default browser
- Direct links to phpMyAdmin and pgweb web admin tools
- Show / hide main window and quit from the tray
- App keeps running in the background when window is closed

### Developer Experience
- Dashboard with service health overview
- Log viewer for Nginx, PHP-FPM, and dnsmasq
- One-time admin password prompt (sudoers-based privilege escalation)
- Shell integration install/uninstall

## Screenshots

| Dashboard | Sites | PHP |
|-----------|-------|-----|
| *coming soon* | *coming soon* | *coming soon* |

## Development

### Requirements

- **macOS** (Apple Silicon or Intel)
- **Go** 1.25+
- **Node.js** 20+ (for frontend)
- **Wails CLI** v2

### Setup

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone & run in dev mode
git clone https://github.com/rizkirmdhnnn/lamboserver.git
cd lamboserver
wails dev
```

### Build

```bash
# Install DMG packaging tool (one-time, local dev only; CI installs this automatically)
# Uses sindresorhus/create-dmg (npm) — zero-config, auto-composes drag-to-Applications visual.
npm install --global create-dmg

# Build .app
wails build -platform darwin/universal -clean

# Build DMG (ad-hoc signed)
./scripts/build-dmg.sh
```

The built app will be at `build/bin/LamboServer.app`.

## Architecture

```
lamboserver/
├── main.go                  # Wails entry point
├── app.go                   # App struct, startup/shutdown, service orchestration
├── internal/
│   ├── system/              # macOS launchd, paths, shell integration
│   ├── process/             # Process runner, PID files, plist generation
│   ├── config/              # Thread-safe JSON config store
│   ├── cert/                # CA and SSL certificate management
│   ├── binaries/            # Binary downloader and registry
│   ├── sites/               # Site linking, Nginx vhost generation
│   ├── tray/                # macOS menu bar tray
│   └── services/
│       ├── nginx/           # Nginx config and lifecycle
│       ├── dnsmasq/         # DNS resolver setup
│       ├── php/             # PHP version management, FPM
│       ├── nodejs/          # Node.js version management
│       ├── postgres/        # PostgreSQL management
│       ├── pgweb/           # pgweb admin interface
│       ├── mysql/           # MySQL management
│       └── phpmyadmin/      # phpMyAdmin interface
├── pkg/
│   ├── logger/              # Structured logger and log reader
│   └── notify/              # macOS notifications
├── scripts/                 # Build and packaging scripts
└── frontend/
    └── src/
        ├── App.tsx          # Tab-based navigation
        └── pages/           # Dashboard, Sites, Services, PHP, Node, Database, Logs
```

### How It Works

1. **First Launch**: Installs a privileged helper script and sudoers entry (one-time admin password prompt)
2. **Service Management**: Nginx and dnsmasq run as macOS LaunchDaemons (root) for privileged port binding. PHP-FPM and databases run as user-level LaunchAgents
3. **Version Switching**: PHP and Node.js versions are managed via symlinks in `~/.lamboserver/bin/`
4. **Site Linking**: Creates Nginx server blocks pointing to your project directory with optional SSL
5. **System Tray**: Menu bar controller bridges to AppKit via CGO (NSStatusBar / NSMenu); closing the window keeps the app alive in the tray
6. **Configuration**: All state persisted in `~/.lamboserver/config.json`

### Data Directory

All runtime data lives in `~/.lamboserver/`:

```
~/.lamboserver/
├── bin/              # Symlinks and helper scripts
├── nginx/            # Nginx config, sites, logs
├── dnsmasq/          # dnsmasq config
├── php/              # Installed PHP versions
├── node/             # Installed Node.js versions
├── postgres/         # PostgreSQL data
├── mysql/            # MySQL data
├── certs/            # CA and site certificates
├── logs/             # Service logs
└── config.json       # App configuration
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Desktop Framework | Wails v2 |
| Backend | Go 1.25 |
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
