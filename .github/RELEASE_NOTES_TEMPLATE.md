## Installing LamboServer on macOS Sequoia

LamboServer is distributed as an **ad-hoc signed** application. It does not carry an Apple Developer ID certificate. macOS Sequoia's Gatekeeper will block the app on first launch with a warning that the app "is damaged or can't be opened."

**This is expected behavior, not a sign of tampering.** Follow one of the steps below to open the app for the first time. Subsequent launches will not show the warning.

### Option 1: Privacy & Security (recommended)

1. Double-click the downloaded `LamboServer.dmg` and drag **LamboServer** to your **Applications** folder.
2. Double-click **LamboServer** in Applications — macOS will show the "damaged" dialog. Click **Cancel** (do not move to Trash).
3. Open **System Settings → Privacy & Security**.
4. Scroll down to the Security section. You will see "LamboServer was blocked to protect your Mac."
5. Click **Open Anyway** and confirm in the follow-up dialog.

### Option 2: Terminal (one command)

If you prefer the command line, clear the quarantine attribute:

```bash
xattr -cr /Applications/LamboServer.app
```

Then double-click **LamboServer** in Applications — it will open without a warning.

### Why this is necessary

LamboServer is an open-source project distributed without an Apple Developer ID ($99/year). Ad-hoc signing is the only option, and macOS treats unsigned / ad-hoc apps with extra caution. This is a Gatekeeper policy, not a defect in the app.
