#ifndef TRAY_DARWIN_H
#define TRAY_DARWIN_H

void CreateTray(const void *iconData, int iconLen, const char *version);
void DestroyTray(void);

#endif
