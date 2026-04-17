# Project Structure — Dev Environment Manager (Wails)

> Panduan struktur folder untuk aplikasi semacam Laragon/XAMPP/Herd yang dibangun dengan Wails.
> Agent harus memastikan setiap file dan folder mengikuti struktur ini saat melakukan refactor.

---

## Services yang Didukung

| Service      | Kategori      | Keterangan                               |
|--------------|---------------|------------------------------------------|
| Nginx        | Web Server    | Reverse proxy & static file serving      |
| DNSMasq      | DNS           | Local DNS resolver & custom domain       |
| PHP          | Runtime       | Multi-version support via PHP-FPM        |
| Node.js      | Runtime       | Multi-version support via binary         |
| MySQL        | Database      | Relational DB                            |
| PostgreSQL   | Database      | Relational DB                            |
| phpMyAdmin   | Web Admin     | Web UI untuk MySQL                       |
| CloudBeaver  | Web Admin     | Web UI untuk PostgreSQL                  |

---

## Hubungan Antar Service

```
Browser
  │
  ├─► :8080  Nginx ──────────────────────────► Static sites / apps
  │           │
  │           ├─► :9081  PHP-FPM 8.1
  │           └─► :9083  PHP-FPM 8.3
  │
  ├─► :8088  phpMyAdmin (served via Nginx + PHP-FPM)
  │                 └─► :3306  MySQL
  │
  └─► :8978  CloudBeaver (standalone Java server)
                    └─► :5432  PostgreSQL

DNSMasq :53 ──► resolve *.test → 127.0.0.1
```

> **Catatan:**
> - phpMyAdmin adalah aplikasi PHP biasa — di-serve oleh Nginx + PHP-FPM yang sudah ada, tidak perlu server terpisah.
> - CloudBeaver adalah aplikasi Java standalone — jalan sebagai server sendiri di port 8978.

---

## Source Code Structure

```
my-dev-env/
├── build/
│   ├── appicon.png
│   ├── darwin/
│   ├── linux/
│   └── windows/
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── common/                     # Button, Modal, Badge, Toggle, StatusDot, dll
│   │   │   ├── dashboard/                  # Overview panel semua services
│   │   │   └── services/
│   │   │       ├── ServiceCard.tsx          # Card generic (nama, status, port, start/stop)
│   │   │       ├── NginxPanel.tsx
│   │   │       ├── DnsmasqPanel.tsx
│   │   │       ├── PhpPanel.tsx
│   │   │       ├── NodejsPanel.tsx
│   │   │       ├── MysqlPanel.tsx
│   │   │       ├── PostgresPanel.tsx
│   │   │       ├── PhpMyAdminPanel.tsx      # Tombol "Open", status, port config
│   │   │       └── CloudBeaverPanel.tsx     # Tombol "Open", status, port config
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Services.tsx
│   │   │   ├── Sites.tsx
│   │   │   └── Settings.tsx
│   │   ├── stores/
│   │   │   ├── serviceStore.ts
│   │   │   └── settingsStore.ts
│   │   └── lib/
│   │       └── wails.ts
│   └── package.json
│
├── internal/
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── defaults.go                     # Termasuk port default phpMyAdmin & CloudBeaver
│   │
│   ├── services/
│   │   ├── service.go                      # Interface Service & VersionedService
│   │   ├── manager.go                      # Orchestrate semua services
│   │   │
│   │   ├── nginx/
│   │   │   ├── nginx.go
│   │   │   ├── installer.go
│   │   │   ├── config.go
│   │   │   └── vhost.go
│   │   │
│   │   ├── dnsmasq/
│   │   │   ├── dnsmasq.go
│   │   │   ├── installer.go
│   │   │   ├── config.go
│   │   │   └── hosts.go
│   │   │
│   │   ├── php/
│   │   │   ├── php.go
│   │   │   ├── installer.go
│   │   │   ├── config.go
│   │   │   └── version.go
│   │   │
│   │   ├── nodejs/
│   │   │   ├── nodejs.go
│   │   │   ├── installer.go
│   │   │   └── version.go
│   │   │
│   │   ├── mysql/
│   │   │   ├── mysql.go
│   │   │   ├── installer.go
│   │   │   ├── config.go
│   │   │   └── version.go
│   │   │
│   │   ├── postgres/
│   │   │   ├── postgres.go
│   │   │   ├── installer.go
│   │   │   ├── config.go
│   │   │   └── version.go
│   │   │
│   │   ├── phpmyadmin/
│   │   │   ├── phpmyadmin.go               # Tidak ada Start/Stop — cukup setup & expose URL
│   │   │   ├── installer.go                # Download & extract phpMyAdmin ke binaries/
│   │   │   └── config.go                   # Generate config.inc.php (host, port, auth)
│   │   │
│   │   └── cloudbeaver/
│   │       ├── cloudbeaver.go              # Start/Stop/Status standalone Java server
│   │       ├── installer.go                # Download & extract CloudBeaver tarball
│   │       └── config.go                   # Generate cloudbeaver.conf & connection preset
│   │
│   ├── process/
│   │   ├── runner.go
│   │   └── port.go
│   │
│   ├── binaries/
│   │   ├── registry.go
│   │   └── downloader.go
│   │
│   ├── sites/
│   │   ├── site.go
│   │   └── manager.go
│   │
│   └── tray/
│       └── tray.go
│
├── pkg/
│   ├── logger/
│   │   └── logger.go
│   └── notify/
│       └── notify.go
│
├── app.go
├── main.go
├── wails.json
└── go.mod
```

---

## Runtime Data Directory (User's Machine)

```
~/.my-dev-env/
├── config.json
│
├── binaries/
│   ├── nginx/1.26.0/
│   ├── dnsmasq/2.90/
│   ├── php/
│   │   ├── 8.1.29/
│   │   └── 8.3.8/
│   ├── nodejs/
│   │   └── 22.4.0/
│   ├── mysql/8.0.38/
│   ├── postgresql/16.3/
│   ├── phpmyadmin/
│   │   └── 5.2.1/                          # Extracted phpMyAdmin files
│   │       ├── index.php
│   │       ├── config.inc.php              # Di-generate oleh app
│   │       └── ...
│   └── cloudbeaver/
│       └── 24.1.0/                         # Extracted CloudBeaver
│           ├── cloudbeaver                 # Executable (Linux/macOS)
│           ├── cloudbeaver.exe             # Executable (Windows)
│           └── ...
│
├── data/
│   ├── mysql/main/
│   └── postgresql/main/
│
├── config/
│   ├── nginx/
│   │   ├── nginx.conf
│   │   └── vhosts/
│   │       ├── myapp.test.conf
│   │       └── phpmyadmin.conf             # Di-generate otomatis saat install phpMyAdmin
│   ├── dnsmasq/dnsmasq.conf
│   ├── php/
│   │   ├── 8.1/php.ini
│   │   ├── 8.1/php-fpm.conf
│   │   └── 8.3/php.ini
│   │   └── 8.3/php-fpm.conf
│   ├── mysql/my.cnf
│   ├── postgresql/postgresql.conf
│   └── cloudbeaver/
│       └── cloudbeaver.conf                # Di-generate oleh app
│
├── logs/
│   ├── nginx.log
│   ├── nginx-error.log
│   ├── dnsmasq.log
│   ├── php-fpm-8.1.log
│   ├── php-fpm-8.3.log
│   ├── mysql.log
│   ├── postgresql.log
│   └── cloudbeaver.log
│
└── tmp/
    ├── nginx.pid
    ├── dnsmasq.pid
    ├── php-fpm-8.1.pid
    ├── php-fpm-8.3.pid
    ├── mysql.pid
    ├── postgresql.pid
    └── cloudbeaver.pid
```

---

## Default Ports

| Service      | Default Port | Keterangan                               |
|--------------|-------------|-------------------------------------------|
| Nginx        | 80, 443     | HTTP & HTTPS                              |
| DNSMasq      | 53          | DNS (butuh privilege di Linux/macOS)      |
| PHP-FPM      | 9081+       | Satu port per versi: 9081, 9082, 9083     |
| Node.js      | —           | Tidak ada daemon; hanya kelola versi PATH |
| MySQL        | 3306        |                                           |
| PostgreSQL   | 5432        |                                           |
| phpMyAdmin   | 8088        | Di-serve via Nginx + PHP-FPM              |
| CloudBeaver  | 8978        | Standalone Java server                    |

---

## Interface Service (`internal/services/service.go`)

```go
type ServiceStatus string

const (
    StatusStopped  ServiceStatus = "stopped"
    StatusStarting ServiceStatus = "starting"
    StatusRunning  ServiceStatus = "running"
    StatusError    ServiceStatus = "error"
)

type Service interface {
    Install(version string) error
    Start() error
    Stop() error
    Restart() error
    Status() ServiceStatus
    Logs() ([]string, error)
    Version() string
}

// Untuk service multi-version aktif (PHP, Node.js)
type VersionedService interface {
    Service
    SwitchVersion(version string) error
    InstalledVersions() []string
    ActiveVersion() string
}

// Untuk web admin yang hanya perlu URL (tidak ada daemon sendiri)
type WebAdminService interface {
    Install(version string) error
    URL() string
    IsInstalled() bool
    Version() string
}
```

| Service      | Interface          | Catatan                                        |
|--------------|--------------------|------------------------------------------------|
| Nginx        | `Service`          |                                                |
| DNSMasq      | `Service`          |                                                |
| PHP          | `VersionedService` |                                                |
| Node.js      | `VersionedService` |                                                |
| MySQL        | `Service`          |                                                |
| PostgreSQL   | `Service`          |                                                |
| phpMyAdmin   | `WebAdminService`  | Tidak ada Start/Stop; di-serve oleh Nginx+PHP  |
| CloudBeaver  | `Service`          | Punya daemon Java sendiri                      |

---

## Wails Bindings (`app.go`)

```go
// ── Generic ───────────────────────────────────────────────────────
func (a *App) GetAllStatuses() map[string]string
func (a *App) StartService(name string) error
func (a *App) StopService(name string) error
func (a *App) RestartService(name string) error
func (a *App) GetServiceLogs(name string, lines int) []string

// ── Version management ────────────────────────────────────────────
func (a *App) GetAvailableVersions(service string) []string
func (a *App) GetInstalledVersions(service string) []string
func (a *App) InstallVersion(service, version string) error
func (a *App) SwitchVersion(service, version string) error

// ── Web Admin ─────────────────────────────────────────────────────
func (a *App) GetWebAdminURL(name string) string       // "phpmyadmin" | "cloudbeaver"
func (a *App) OpenWebAdmin(name string) error          // Buka di browser default
func (a *App) InstallWebAdmin(name, version string) error

// ── Sites ─────────────────────────────────────────────────────────
func (a *App) GetSites() []Site
func (a *App) CreateSite(domain, path, phpVersion string) error
func (a *App) DeleteSite(domain string) error

// ── Config ────────────────────────────────────────────────────────
func (a *App) GetConfig() AppConfig
func (a *App) SaveConfig(cfg AppConfig) error
```

---

## Key Principles untuk Agent

| # | Prinsip | Detail |
|---|---------|--------|
| 1 | **Jangan bundle binary** | Semua binary di-download on-demand ke `~/.my-dev-env/binaries/` |
| 2 | **Interface wajib diikuti** | Lihat tabel interface di atas; jangan implement Start/Stop di phpMyAdmin |
| 3 | **phpMyAdmin = static files** | Di-serve oleh Nginx + PHP-FPM yang sudah ada; app hanya generate `config.inc.php` & vhost Nginx |
| 4 | **CloudBeaver = daemon sendiri** | Jalan sebagai Java process terpisah; butuh JRE tersedia di sistem |
| 5 | **app.go hanya glue** | Tidak ada business logic di `app.go` |
| 6 | **Wails events untuk realtime** | Gunakan `runtime.EventsEmit` untuk push status |
| 7 | **Cross-platform paths** | Selalu `os.UserConfigDir()`, `filepath.Join()` |
| 8 | **Sites = Nginx + DNS** | Create site → generate vhost + DNS entry + reload keduanya |
| 9 | **PHP multi-version via FPM** | Setiap versi PHP jalan di port FPM tersendiri |
| 10 | **OpenWebAdmin via browser** | Gunakan `browser.OpenURL()` dari Wails runtime |

---

## Cara Menambah Service Baru

1. Buat folder `internal/services/<nama>/`
2. Tentukan interface yang sesuai (`Service`, `VersionedService`, atau `WebAdminService`)
3. Buat minimal: `<nama>.go`, `installer.go`, `config.go`
4. Daftarkan di `internal/services/manager.go`
5. Tambahkan binding di `app.go`
6. Tambahkan default port di `internal/config/defaults.go`
7. Buat komponen UI `frontend/src/components/services/<Nama>Panel.tsx`

---

*File ini dibuat sebagai referensi refactoring. Agent harus menggunakan dokumen ini sebagai satu-satunya acuan saat memindahkan, membuat, atau mengganti nama file.*
