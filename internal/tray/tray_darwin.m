#import <Cocoa/Cocoa.h>
#include "tray_darwin.h"

// Forward declarations for Go callbacks
extern void onShowWindow(void);
extern void onQuit(void);
extern void onServiceAction(const char *name, const char *action);
extern void onRefreshStatuses(void);

// Service descriptors: registry key, display name (D-02 dependency order)
static NSString *serviceKeys[] = {@"dnsmasq", @"nginx", @"php", @"mysql", @"postgresql"};
static NSString *serviceNames[] = {@"DNSMasq", @"Nginx", @"PHP", @"MySQL", @"PostgreSQL"};
static const int kServiceCount = 5;

// LamboSystrayDelegate -- unique name to avoid Wails ObjC symbol collisions (D-02)
@interface LamboSystrayDelegate : NSObject <NSMenuDelegate>
@property (strong, nonatomic) NSStatusItem *statusItem;
@property (strong, nonatomic) NSMutableArray<NSMenuItem *> *serviceItems;
@property (strong, nonatomic) NSMutableArray<NSMenu *> *serviceSubmenus;
@end

static LamboSystrayDelegate *delegate = nil;

@implementation LamboSystrayDelegate

- (void)showWindow:(id)sender {
    onShowWindow();
}

- (void)quit:(id)sender {
    onQuit();
}

- (void)serviceStart:(id)sender {
    int idx = (int)[sender tag];
    onServiceAction([serviceKeys[idx] UTF8String], "start");
}

- (void)serviceStop:(id)sender {
    int idx = (int)[sender tag];
    onServiceAction([serviceKeys[idx] UTF8String], "stop");
}

- (void)serviceRestart:(id)sender {
    int idx = (int)[sender tag];
    onServiceAction([serviceKeys[idx] UTF8String], "restart");
}

// NSMenuDelegate: refresh statuses before menu displays (D-07)
- (void)menuWillOpen:(NSMenu *)menu {
    // Synchronous call into Go. The CGO call chain is:
    // menuWillOpen: (main thread) -> onRefreshStatuses (Go) -> UpdateServiceStatus (C)
    // All on the same thread, so updates apply before menu renders.
    onRefreshStatuses();
}

- (void)updateServiceItem:(int)index withStatus:(NSString *)status {
    if (index < 0 || index >= kServiceCount) return;

    NSMenuItem *item = self.serviceItems[index];
    NSMenu *submenu = self.serviceSubmenus[index];
    NSString *name = serviceNames[index];

    if ([status isEqualToString:@"not_installed"]) {
        // D-08: dash, "Not Installed" label, disabled, no submenu
        NSMutableAttributedString *title = [[NSMutableAttributedString alloc] init];
        NSDictionary *dotAttrs = @{
            NSForegroundColorAttributeName: [NSColor tertiaryLabelColor],
            NSFontAttributeName: [NSFont menuFontOfSize:14.0]
        };
        NSDictionary *nameAttrs = @{
            NSFontAttributeName: [NSFont menuFontOfSize:14.0]
        };
        [title appendAttributedString:[[NSAttributedString alloc]
            initWithString:@"\u2500 " attributes:dotAttrs]];
        [title appendAttributedString:[[NSAttributedString alloc]
            initWithString:[NSString stringWithFormat:@"%@ \u2014 Not Installed", name]
                attributes:nameAttrs]];
        [item setAttributedTitle:title];
        [item setEnabled:NO];
        [item.menu setSubmenu:nil forItem:item];
        return;
    }

    // Installed service: show colored dot
    BOOL isRunning = [status isEqualToString:@"running"];
    NSString *dot = isRunning ? @"\u25CF " : @"\u25CB ";
    NSColor *dotColor = isRunning
        ? [NSColor systemGreenColor]
        : [NSColor secondaryLabelColor];

    NSMutableAttributedString *title = [[NSMutableAttributedString alloc] init];
    NSDictionary *dotAttrs = @{
        NSForegroundColorAttributeName: dotColor,
        NSFontAttributeName: [NSFont menuFontOfSize:14.0]
    };
    NSDictionary *nameAttrs = @{
        NSFontAttributeName: [NSFont menuFontOfSize:14.0]
    };
    [title appendAttributedString:[[NSAttributedString alloc]
        initWithString:dot attributes:dotAttrs]];
    [title appendAttributedString:[[NSAttributedString alloc]
        initWithString:name attributes:nameAttrs]];
    [item setAttributedTitle:title];
    [item setEnabled:YES];

    // Re-attach submenu (may have been removed if previously not_installed)
    if (item.submenu == nil) {
        [item.menu setSubmenu:submenu forItem:item];
    }

    // D-05: Context-aware submenu actions
    NSArray<NSMenuItem *> *subItems = submenu.itemArray;
    // subItems[0] = Start, subItems[1] = Stop, subItems[2] = Restart
    [subItems[0] setEnabled:!isRunning];  // Start disabled when running
    [subItems[1] setEnabled:isRunning];   // Stop disabled when stopped
    [subItems[2] setEnabled:isRunning];   // Restart disabled when stopped
}

@end

void CreateTray(const void *iconData, int iconLen, const char *version) {
    dispatch_async(dispatch_get_main_queue(), ^{
        delegate = [[LamboSystrayDelegate alloc] init];

        NSStatusBar *bar = [NSStatusBar systemStatusBar];
        delegate.statusItem = [bar statusItemWithLength:NSVariableStatusItemLength];

        NSData *data = [NSData dataWithBytes:iconData length:iconLen];
        NSImage *icon = [[NSImage alloc] initWithData:data];
        [icon setTemplate:YES];  // D-04: enables dark/light mode adaptation
        [icon setSize:NSMakeSize(22, 22)];
        delegate.statusItem.button.image = icon;

        // Build menu per D-05/D-06
        NSMenu *menu = [[NSMenu alloc] init];

        // Header: "LamboServer v1.0.0" (disabled)
        NSString *headerTitle = [NSString stringWithFormat:@"LamboServer v%s", version];
        NSMenuItem *header = [[NSMenuItem alloc] initWithTitle:headerTitle
                                                        action:nil
                                                 keyEquivalent:@""];
        [header setEnabled:NO];
        [menu addItem:header];

        [menu addItem:[NSMenuItem separatorItem]];

        // Service items per D-01, D-02 (dependency order)
        delegate.serviceItems = [NSMutableArray arrayWithCapacity:kServiceCount];
        delegate.serviceSubmenus = [NSMutableArray arrayWithCapacity:kServiceCount];

        for (int i = 0; i < kServiceCount; i++) {
            NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:serviceNames[i]
                                                          action:nil
                                                   keyEquivalent:@""];

            // Create submenu with Start/Stop/Restart per D-03
            NSMenu *sub = [[NSMenu alloc] init];

            NSMenuItem *startItem = [[NSMenuItem alloc] initWithTitle:@"Start"
                                                               action:@selector(serviceStart:)
                                                        keyEquivalent:@""];
            [startItem setTarget:delegate];
            [startItem setTag:i];
            [sub addItem:startItem];

            NSMenuItem *stopItem = [[NSMenuItem alloc] initWithTitle:@"Stop"
                                                              action:@selector(serviceStop:)
                                                       keyEquivalent:@""];
            [stopItem setTarget:delegate];
            [stopItem setTag:i];
            [sub addItem:stopItem];

            NSMenuItem *restartItem = [[NSMenuItem alloc] initWithTitle:@"Restart"
                                                                 action:@selector(serviceRestart:)
                                                          keyEquivalent:@""];
            [restartItem setTarget:delegate];
            [restartItem setTag:i];
            [sub addItem:restartItem];

            [delegate.serviceItems addObject:item];
            [delegate.serviceSubmenus addObject:sub];

            [menu addItem:item];
            [menu setSubmenu:sub forItem:item];
        }

        [menu addItem:[NSMenuItem separatorItem]];

        // Show Window
        NSMenuItem *showItem = [[NSMenuItem alloc]
            initWithTitle:@"Show Window"
                   action:@selector(showWindow:)
            keyEquivalent:@""];
        [showItem setTarget:delegate];
        [menu addItem:showItem];

        [menu addItem:[NSMenuItem separatorItem]];

        // Quit
        NSMenuItem *quitItem = [[NSMenuItem alloc]
            initWithTitle:@"Quit"
                   action:@selector(quit:)
            keyEquivalent:@""];
        [quitItem setTarget:delegate];
        [menu addItem:quitItem];

        delegate.statusItem.menu = menu;
        menu.delegate = delegate;
    });
}

void DestroyTray(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (delegate != nil && delegate.statusItem != nil) {
            [[NSStatusBar systemStatusBar] removeStatusItem:delegate.statusItem];
            delegate.statusItem = nil;
        }
        delegate = nil;
    });
}

// UpdateServiceStatus is called from Go (onRefreshStatuses) for each service.
// The call chain is synchronous on the main thread (menuWillOpen: -> Go -> here),
// so no GCD dispatch is needed.
void UpdateServiceStatus(int index, const char *status) {
    NSString *statusStr = [NSString stringWithUTF8String:status];
    [delegate updateServiceItem:index withStatus:statusStr];
}

// RefreshServiceStatuses is a convenience C wrapper around the Go export.
void RefreshServiceStatuses(void) {
    onRefreshStatuses();
}
