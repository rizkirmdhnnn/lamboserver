# Phase 3: Quick Access Links - Pattern Map

**Mapped:** 2026-04-18
**Files analyzed:** 4 modified files
**Analogs found:** 4 / 4

---

## File Classification

| Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---------------|------|-----------|----------------|---------------|
| `internal/tray/controller.go` | interface definition | request-response | itself (Phase 2 extension) | exact |
| `internal/tray/controller_darwin.go` | CGO bridge / callback | request-response | itself (Phase 2 `onRefreshStatuses`, `onServiceAction`) | exact |
| `internal/tray/tray_darwin.h` | C header / config | request-response | itself (Phase 2 declarations) | exact |
| `internal/tray/tray_darwin.m` | ObjC UI / event-driven | event-driven | itself (Phase 2 `menuWillOpen:`, `updateServiceItem:`) | exact |
| `app.go` | controller / service layer | request-response | `app.go` `OpenWebAdmin`, `GetSites` (lines 442–456, 670–672) | exact |

---

## Pattern Assignments

### `internal/tray/controller.go` (interface extension)

**Analog:** itself — current content at `/Users/rizkirmdhn/Documents/Code/lamboserver/internal/tray/controller.go`

**Existing interface pattern** (lines 7–18 — copy and extend):
```go
// AppController is the narrow interface the tray needs from the application.
// Follows consumer-defined interface convention (D-02).
type AppController interface {
    Context() context.Context
    GetAllStatuses() map[string]string
    StartService(name string) error
    StopService(name string) error
    RestartService(name string) error
}
```

**New methods to add to AppController** (append to interface body):
```go
    // GetTraySites returns sorted (newest-first), capped site info for the tray menu.
    GetTraySites() []SiteInfo
    // OpenSiteInBrowser opens the site URL (https or http per SSLEnabled) in the default browser.
    OpenSiteInBrowser(domain string)
    // GetWebAdminItems returns install status and URL for phpMyAdmin and pgweb.
    GetWebAdminItems() []WebAdminItem
    // OpenWebAdmin opens the named web admin tool (phpmyadmin or pgweb) in the default browser.
    OpenWebAdmin(name string)
    // GetTotalSiteCount returns total site count for overflow computation.
    GetTotalSiteCount() int
```

**New value types to define in controller.go** (after the interface — no domain package imports):
```go
// SiteInfo is the tray-facing site descriptor. Domain-only label (Claude's discretion).
type SiteInfo struct {
    Domain string
    URL    string // Pre-computed: "https://domain" or "http://domain"
    Label  string // Display string — domain only, e.g. "myapp.test"
}

// WebAdminItem is the tray-facing descriptor for a web admin tool.
type WebAdminItem struct {
    Name      string // Registry key: "phpmyadmin" or "pgweb"
    Label     string // Display: "Open phpMyAdmin" / "Open pgweb"
    Installed bool
    URL       string
}
```

---

### `internal/tray/controller_darwin.go` (CGO bridge extension)

**Analog:** itself — `/Users/rizkirmdhn/Documents/Code/lamboserver/internal/tray/controller_darwin.go`

**Guard pattern** (lines 39–48 — copy for every new `//export` function):
```go
//export onShowWindow
func onShowWindow() {
    if instance == nil || instance.app == nil {
        return
    }
    ctx := instance.app.Context()
    ...
}
```

**C string + fire-and-forget pattern** (lines 63–80 — copy for `onOpenSite` and `onOpenWebAdmin`):
```go
//export onServiceAction
func onServiceAction(cName *C.char, cAction *C.char) {
    if instance == nil || instance.app == nil {
        return
    }
    name := C.GoString(cName)
    action := C.GoString(cAction)
    go func() {
        switch action {
        case "start":
            instance.app.StartService(name)
        ...
        }
    }()
}
```

**C string slice + free pattern** (lines 83–99 — copy for `onRefreshQuickAccess`):
```go
//export onRefreshStatuses
func onRefreshStatuses() {
    if instance == nil || instance.app == nil {
        return
    }
    statuses := instance.app.GetAllStatuses()
    serviceOrder := []string{"dnsmasq", "nginx", "php", "mysql", "postgresql"}
    for i, name := range serviceOrder {
        status, ok := statuses[name]
        if !ok {
            status = "not_installed"
        }
        cStatus := C.CString(status)
        C.UpdateServiceStatus(C.int(i), cStatus)
        C.free(unsafe.Pointer(cStatus))  // free immediately after C call
    }
}
```

**New callbacks to add — follow exact same guard + goroutine + free structure:**
```go
//export onOpenSite
func onOpenSite(cDomain *C.char) {
    if instance == nil || instance.app == nil {
        return
    }
    domain := C.GoString(cDomain)
    go func() {
        instance.app.OpenSiteInBrowser(domain)
    }()
}

//export onOpenWebAdmin
func onOpenWebAdmin(cName *C.char) {
    if instance == nil || instance.app == nil {
        return
    }
    name := C.GoString(cName)
    go func() {
        instance.app.OpenWebAdmin(name)
    }()
}

//export onRefreshQuickAccess
func onRefreshQuickAccess() {
    if instance == nil || instance.app == nil {
        return
    }
    sites := instance.app.GetTraySites()
    // Pass count and overflow hint to ObjC before individual items
    C.BeginQuickAccessRebuild(C.int(len(sites)))
    for i, s := range sites {
        cDomain := C.CString(s.Domain)
        cURL := C.CString(s.URL)
        cLabel := C.CString(s.Label)
        C.AddQuickAccessSite(C.int(i), cDomain, cURL, cLabel)
        C.free(unsafe.Pointer(cDomain))
        C.free(unsafe.Pointer(cURL))
        C.free(unsafe.Pointer(cLabel))
    }
    // Compute and pass overflow count
    overflow := instance.app.GetTotalSiteCount() - len(sites)
    if overflow < 0 {
        overflow = 0
    }
    C.SetQuickAccessOverflow(C.int(overflow))
    webAdmins := instance.app.GetWebAdminItems()
    for _, wa := range webAdmins {
        if !wa.Installed {
            continue
        }
        cName := C.CString(wa.Name)
        cLabel := C.CString(wa.Label)
        C.AddQuickAccessWebAdmin(cName, cLabel)
        C.free(unsafe.Pointer(cName))
        C.free(unsafe.Pointer(cLabel))
    }
    C.CommitQuickAccessRebuild()
}
```

---

### `internal/tray/tray_darwin.h` (C header extension)

**Analog:** itself — `/Users/rizkirmdhn/Documents/Code/lamboserver/internal/tray/tray_darwin.h`

**Existing declaration pattern** (lines 1–9):
```c
#ifndef TRAY_DARWIN_H
#define TRAY_DARWIN_H

void CreateTray(const void *iconData, int iconLen, const char *version);
void DestroyTray(void);
void UpdateServiceStatus(int index, const char *status);
void RefreshServiceStatuses(void);

#endif
```

**New declarations to add** (insert before `#endif`):
```c
// Quick Access submenu data-passing protocol (Phase 3)
void BeginQuickAccessRebuild(int siteCount);
void AddQuickAccessSite(int index, const char *domain, const char *url, const char *label);
void SetQuickAccessOverflow(int count);
void AddQuickAccessWebAdmin(const char *name, const char *label);
void CommitQuickAccessRebuild(void);
```

---

### `internal/tray/tray_darwin.m` (ObjC menu extension)

**Analog:** itself — `/Users/rizkirmdhn/Documents/Code/lamboserver/internal/tray/tray_darwin.m`

**Forward declaration pattern** (lines 5–8 — add new callbacks alongside existing):
```objc
extern void onShowWindow(void);
extern void onQuit(void);
extern void onServiceAction(const char *name, const char *action);
extern void onRefreshStatuses(void);
// New for Phase 3:
// extern void onOpenSite(const char *domain);
// extern void onOpenWebAdmin(const char *name);
// extern void onRefreshQuickAccess(void);
```

**Interface property pattern** (lines 16–20 — add new properties to LamboSystrayDelegate):
```objc
@interface LamboSystrayDelegate : NSObject <NSMenuDelegate>
@property (strong, nonatomic) NSStatusItem *statusItem;
@property (strong, nonatomic) NSMutableArray<NSMenuItem *> *serviceItems;
@property (strong, nonatomic) NSMutableArray<NSMenu *> *serviceSubmenus;
// New for Phase 3:
// @property (strong, nonatomic) NSMenu *quickAccessSubmenu;
// @property (strong, nonatomic) NSMutableArray<NSDictionary *> *quickAccessSites;
// @property (assign, nonatomic) int quickAccessOverflow;
// @property (strong, nonatomic) NSMutableArray<NSDictionary *> *quickAccessWebAdmins;
@end
```

**menuWillOpen: pattern** (lines 50–55 — append new refresh call):
```objc
- (void)menuWillOpen:(NSMenu *)menu {
    onRefreshStatuses();      // existing
    onRefreshQuickAccess();   // new: rebuilds Quick Access submenu data + calls rebuildQuickAccessSubmenu
}
```

**NSMenuItem + tag action pattern** (lines 34–47 — copy for openSite: and openWebAdmin:):
```objc
- (void)serviceStart:(id)sender {
    int idx = (int)[sender tag];
    onServiceAction([serviceKeys[idx] UTF8String], "start");
}
```

**Action methods to add** (follow same ObjC delegation pattern):
```objc
- (void)openSite:(id)sender {
    NSString *domain = [sender representedObject];
    onOpenSite([domain UTF8String]);
}

- (void)openWebAdmin:(id)sender {
    NSString *name = [sender representedObject];
    onOpenWebAdmin([name UTF8String]);
}
```

**submenu rebuild pattern** (rebuild from scratch per pitfall 3 — use `removeAllItems`):
```objc
- (void)rebuildQuickAccessSubmenu {
    [self.quickAccessSubmenu removeAllItems];

    BOOL hasSites = self.quickAccessSites.count > 0;
    BOOL hasWebAdmin = self.quickAccessWebAdmins.count > 0;

    if (!hasSites && !hasWebAdmin) {
        NSMenuItem *empty = [[NSMenuItem alloc]
            initWithTitle:@"No sites configured" action:nil keyEquivalent:@""];
        [empty setEnabled:NO];
        [self.quickAccessSubmenu addItem:empty];
        return;
    }

    for (NSDictionary *site in self.quickAccessSites) {
        NSMenuItem *item = [[NSMenuItem alloc]
            initWithTitle:site[@"label"]
                   action:@selector(openSite:)
            keyEquivalent:@""];
        [item setTarget:self];
        [item setRepresentedObject:site[@"domain"]];
        [self.quickAccessSubmenu addItem:item];
    }

    if (self.quickAccessOverflow > 0) {
        NSString *title = [NSString stringWithFormat:
            @"(+%d more \u2014 see main window)", self.quickAccessOverflow];
        NSMenuItem *overflow = [[NSMenuItem alloc]
            initWithTitle:title action:nil keyEquivalent:@""];
        [overflow setEnabled:NO];
        [self.quickAccessSubmenu addItem:overflow];
    }

    if (hasSites && hasWebAdmin) {
        [self.quickAccessSubmenu addItem:[NSMenuItem separatorItem]];
    }
    // Web admin items added here from quickAccessWebAdmins array
    for (NSDictionary *wa in self.quickAccessWebAdmins) {
        NSMenuItem *item = [[NSMenuItem alloc]
            initWithTitle:wa[@"label"]
                   action:@selector(openWebAdmin:)
            keyEquivalent:@""];
        [item setTarget:self];
        [item setRepresentedObject:wa[@"name"]];
        [self.quickAccessSubmenu addItem:item];
    }
}
```

**GCD dispatch + submenu insertion in CreateTray** (lines 122–210 — insert Quick Access between services separator and Show Window):
```objc
// After services loop and separator (around line 188):
[menu addItem:[NSMenuItem separatorItem]]; // existing separator after services

// Quick Access parent item + submenu (D-02: between services and Show Window)
NSMenuItem *qaItem = [[NSMenuItem alloc]
    initWithTitle:@"Quick Access" action:nil keyEquivalent:@""];
delegate.quickAccessSubmenu = [[NSMenu alloc] init];
[qaItem setSubmenu:delegate.quickAccessSubmenu];
[menu addItem:qaItem];

[menu addItem:[NSMenuItem separatorItem]]; // separator before Show Window

// Show Window (existing)
```

---

### `app.go` (new methods)

**Analog:** `app.go` lines 442–456 (`OpenWebAdmin`) and lines 670–672 (`GetSites`)

**Delegation + browser.OpenURL pattern** (lines 442–456):
```go
func (a *App) OpenWebAdmin(name string) error {
    a.Debug.Info("OpenWebAdmin called: %s", name)
    if svc, err := a.Manager.GetWebAdmin(name); err == nil {
        url := svc.URL()
        if url == "" {
            return fmt.Errorf("web admin %q has no URL", name)
        }
        return browser.OpenURL(url)
    }
    return fmt.Errorf("web admin %q not found", name)
}
```

**New methods to add to app.go — follow delegation + Debug.Info pattern:**

`GetTraySites() []tray.SiteInfo` — sort newest-first (D-04), cap at 15 (D-05), build URL from SSLEnabled (D-03):
```go
// GetTraySites returns the sorted, capped site list for the tray package.
// Satisfies tray.AppController.GetTraySites (D-04: newest first, D-05: cap at 15).
// Named GetTraySites (not GetSites) to avoid collision with existing Wails-bound GetSites() []sites.Site.
func (a *App) GetTraySites() []tray.SiteInfo {
    raw := a.Sites.List()
    sort.Slice(raw, func(i, j int) bool {
        return raw[i].CreatedAt > raw[j].CreatedAt // newest first; RFC3339 is lexicographically sortable
    })
    const maxSites = 15
    if len(raw) > maxSites {
        raw = raw[:maxSites]
    }
    items := make([]tray.SiteInfo, len(raw))
    for i, s := range raw {
        scheme := "https"
        if !s.SSLEnabled {
            scheme = "http"
        }
        items[i] = tray.SiteInfo{
            Domain: s.Domain,
            URL:    scheme + "://" + s.Domain,
            Label:  s.Domain,
        }
    }
    return items
}
```

`GetTotalSiteCount() int` — returns total count for overflow computation:
```go
// GetTotalSiteCount returns the total number of configured sites.
// Used by tray to compute overflow count for the "(+N more)" label.
func (a *App) GetTotalSiteCount() int {
    return len(a.Sites.List())
}
```

`OpenSiteInBrowser(domain string)` — construct URL and call browser.OpenURL:
```go
func (a *App) OpenSiteInBrowser(domain string) {
    a.Debug.Info("OpenSiteInBrowser called: %s", domain)
    // Re-derive the URL from config so a deleted site gracefully fails silently.
    for _, s := range a.Sites.List() {
        if s.Domain == domain {
            scheme := "https"
            if !s.SSLEnabled {
                scheme = "http"
            }
            _ = browser.OpenURL(scheme + "://" + domain)
            return
        }
    }
    // Domain no longer in config — open https as best-effort.
    _ = browser.OpenURL("https://" + domain)
}
```

`GetWebAdminItems() []tray.WebAdminItem` — bridge phpmyadmin registry + pgweb direct field:
```go
// GetWebAdminItems returns install status for phpMyAdmin (via registry) and pgweb
// (via direct Manager field — pgweb is NOT registered as a WebAdminService).
func (a *App) GetWebAdminItems() []tray.WebAdminItem {
    var items []tray.WebAdminItem
    if pma, err := a.Manager.GetWebAdmin("phpmyadmin"); err == nil {
        items = append(items, tray.WebAdminItem{
            Name:      "phpmyadmin",
            Label:     "Open phpMyAdmin",
            Installed: pma.IsInstalled(),
            URL:       pma.URL(),
        })
    }
    items = append(items, tray.WebAdminItem{
        Name:      "pgweb",
        Label:     "Open pgweb",
        Installed: a.Pgweb.IsInstalled(),
        URL:       a.Pgweb.URL(),
    })
    return items
}
```

`OpenWebAdmin(name string)` for tray (same pattern as existing `OpenWebAdmin` but handles pgweb too):
```go
// The existing OpenWebAdmin in app.go only covers the WebAdmin registry (phpmyadmin).
// The tray calls instance.app.OpenWebAdmin(name) which must also handle "pgweb".
// Extend the existing method by adding a pgweb fallback before the final error return:
    if name == "pgweb" {
        a.Debug.Info("OpenWebAdmin: pgweb via direct manager")
        return browser.OpenURL(a.Pgweb.URL())
    }
```

---

## Shared Patterns

### GCD Main Queue Dispatch
**Source:** `internal/tray/tray_darwin.m` lines 123, 214
**Apply to:** All Cocoa UI mutations in `CreateTray`, `DestroyTray`, and `CommitQuickAccessRebuild`
```objc
dispatch_async(dispatch_get_main_queue(), ^{
    // all NSMenu / NSMenuItem mutations here
});
```

### CGO Guard + Goroutine for Action Callbacks
**Source:** `internal/tray/controller_darwin.go` lines 63–80
**Apply to:** `onOpenSite`, `onOpenWebAdmin`
```go
if instance == nil || instance.app == nil {
    return
}
// Always fire browser launch in a goroutine — never block ObjC main thread
go func() {
    instance.app.SomeAction(param)
}()
```

### C String Free Pattern
**Source:** `internal/tray/controller_darwin.go` lines 83–99
**Apply to:** `onRefreshQuickAccess` — every `C.CString()` call must be followed by `C.free(unsafe.Pointer(ptr))`
```go
cStr := C.CString(goStr)
C.SomeCFunction(cStr)
C.free(unsafe.Pointer(cStr))
```

### NSMenuItem setRepresentedObject: for String Data
**Source:** `internal/tray/tray_darwin.m` lines 161–186 (tag pattern for services)
**Apply to:** Site items and web admin items where a string identifier (domain, name) must survive the click action
```objc
[item setRepresentedObject:@"phpmyadmin"];
// In action method:
NSString *name = [sender representedObject];
```

### Debug.Info Logging for App Methods
**Source:** `app.go` line 444 (`a.Debug.Info("OpenWebAdmin called: %s", name)`)
**Apply to:** All new public methods on `App` — log at entry with method name and key params
```go
a.Debug.Info("MethodName called: param=%s", param)
```

---

## No Analog Found

All modified files have exact analogs (themselves at their Phase 2 state). No files require pattern invention.

---

## Metadata

**Analog search scope:** `internal/tray/`, `app.go`, `internal/sites/manager.go`, `internal/services/pgweb/manager.go`, `internal/services/phpmyadmin/manager.go`
**Files scanned:** 9
**Pattern extraction date:** 2026-04-18
