---
phase: 06-add-pgweb
asvs_level: 1
audited: 2026-04-16
result: SECURED
threats_total: 6
threats_closed: 6
threats_open: 0
---

# Phase 06 Security Audit

## Summary

All 6 threats in the threat register are closed. Both mitigated threats have confirmed code evidence. All 4 accepted risks are documented below. No unregistered threat flags were raised by either plan executor.

---

## Threat Verification

| Threat ID | Category | Disposition | Status | Evidence |
|-----------|----------|-------------|--------|----------|
| T-06-01 | Elevation of Privilege | mitigate | CLOSED | `internal/services/pgweb/manager.go:139` — `"--bind=127.0.0.1"` passed to every `exec.Command` invocation of pgweb |
| T-06-02 | Denial of Service | mitigate | CLOSED | `internal/services/pgweb/manager.go:113-120` — `checkPort()` performs `net.Listen("tcp", "127.0.0.1:8081")` preflight before `Start()` proceeds; returns descriptive error on conflict |
| T-06-03 | Tampering | accept | CLOSED | See accepted risks log below |
| T-06-04 | Information Disclosure | accept | CLOSED | See accepted risks log below |
| T-06-05 | Spoofing | accept | CLOSED | See accepted risks log below |
| T-06-06 | Information Disclosure | accept | CLOSED | See accepted risks log below |

---

## Mitigated Threats — Code Evidence

### T-06-01: Elevation of Privilege — `--bind=127.0.0.1` always set

File: `internal/services/pgweb/manager.go`, lines 136-142

```go
cmd := exec.Command(m.binaryPath(),
    "--host=127.0.0.1",
    "--user=postgres",
    "--bind=127.0.0.1",   // D-14: always bind to loopback only
    "--listen=8081",
    "--skip-open",
)
```

The `--bind=127.0.0.1` flag is hardcoded and unconditional. pgweb cannot be reached from the LAN regardless of OS firewall configuration.

### T-06-02: Denial of Service — `checkPort()` preflight

File: `internal/services/pgweb/manager.go`, lines 113-134

```go
func (m *Manager) checkPort() error {
    ln, err := net.Listen("tcp", "127.0.0.1:8081")
    if err != nil {
        return fmt.Errorf("port 8081 is already in use by another process")
    }
    ln.Close()
    return nil
}

func (m *Manager) Start() error {
    // ...
    if err := m.checkPort(); err != nil {
        return err
    }
    // ...
}
```

Port conflict is detected before the pgweb binary is launched. The returned error is descriptive and surfaces to the frontend via the `StartPgweb` IPC method.

---

## Accepted Risks Log

### T-06-03: Tampering — pgweb binary download

- **Threat:** Downloaded binary could be tampered with in transit or at the GitHub release endpoint.
- **Rationale:** Download uses HTTPS (`https://github.com/sosedoff/pgweb/releases/download/...`), providing transport integrity. pgweb is a local-development-only tool with no network-exposed attack surface. The binary is a low-value target. No checksum verification is performed post-download.
- **Residual risk:** Low. Accepted for local-dev tooling.
- **Evidence:** `internal/services/pgweb/interfaces.go:12` — HTTPS URL hardcoded.

### T-06-04: Information Disclosure — PostgreSQL trust auth

- **Threat:** pgweb connects to PostgreSQL using trust authentication (no password), meaning any local process can connect as `postgres`.
- **Rationale:** Trust authentication is the standard configuration for local development PostgreSQL instances. The PostgreSQL socket and port 5432 are bound to localhost only. This is consistent with the existing PostgreSQL configuration used by the rest of the application.
- **Residual risk:** Low within a single-user developer workstation. Not suitable for shared or multi-tenant environments.
- **Evidence:** `internal/services/pgweb/manager.go:138` — `"--user=postgres"` connects without password.

### T-06-05: Spoofing — Wails IPC calls

- **Threat:** Frontend Wails IPC calls to pgweb methods could be spoofed or replayed by a malicious process.
- **Rationale:** Wails IPC is implemented via a local WebView bridge, not a network socket. There is no network exposure. Only the host process (the app itself) can invoke IPC methods.
- **Residual risk:** None in normal operation. Accepted.
- **Evidence:** `app.go:484-527` — pgweb IPC methods are standard Wails public methods with no additional auth layer needed.

### T-06-06: Information Disclosure — pgweb URL in frontend

- **Threat:** The pgweb localhost URL (port 8081) is embedded in frontend code and could reveal internal topology.
- **Rationale:** The URL `http://127.0.0.1:8081` is display-only text in the running state label. It references only a loopback address and port number — no credentials, tokens, or sensitive path information are exposed. This data has no value to an attacker.
- **Residual risk:** None. Accepted.
- **Evidence:** `frontend/src/pages/PgDatabasePage.tsx:337` — `"Running \u00B7 Port 8081"` display string only.

---

## Unregistered Threat Flags

None. Both plan summaries (`06-01-SUMMARY.md`, `06-02-SUMMARY.md`) report `## Threat Flags: None`.

---

## Audit Notes

- ASVS Level 1 applied. No additional controls required at this level for a local-development-only tool.
- The `block_on: high` policy is satisfied — no open threats exist.
- Implementation files were read-only during this audit. No implementation changes were made.
