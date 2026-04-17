#import <Cocoa/Cocoa.h>
#include "tray_darwin.h"

// Forward declarations for Go callbacks
extern void onShowWindow(void);
extern void onQuit(void);

// LamboSystrayDelegate -- unique name to avoid Wails ObjC symbol collisions (D-02)
@interface LamboSystrayDelegate : NSObject
@property (strong, nonatomic) NSStatusItem *statusItem;
@end

static LamboSystrayDelegate *delegate = nil;

@implementation LamboSystrayDelegate

- (void)showWindow:(id)sender {
    onShowWindow();
}

- (void)quit:(id)sender {
    onQuit();
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
