#ifndef TRAY_DARWIN_H
#define TRAY_DARWIN_H

void CreateTray(const void *iconData, int iconLen, const char *version);
void DestroyTray(void);
void UpdateServiceStatus(int index, const char *status);
void RefreshServiceStatuses(void);

// Quick Access submenu data-passing protocol (Phase 3)
void BeginQuickAccessRebuild(int siteCount);
void AddQuickAccessSite(int index, const char *domain, const char *url, const char *label);
void SetQuickAccessOverflow(int count);
void AddQuickAccessWebAdmin(const char *name, const char *label);
void CommitQuickAccessRebuild(void);

#endif
