# LamboServer v1.0.0

## Download

Download `LamboServer-1.0.0.dmg`, open it, and drag LamboServer to your Applications folder.

## First Launch on macOS

LamboServer is not notarized by Apple. On first launch, macOS will block the app.

**To open LamboServer:**

1. Open **System Settings** > **Privacy & Security**
2. Scroll down to the Security section
3. You will see "LamboServer was blocked from use because it is not from an identified developer"
4. Click **Open Anyway**
5. Enter your password when prompted

**Alternative (Terminal):**

```bash
xattr -r -d com.apple.quarantine /Applications/LamboServer.app
```

This warning appears only once per installation.

## What's Included

- LamboServer desktop app (universal binary -- Apple Silicon and Intel)
- Manage Nginx, PHP, MySQL, PostgreSQL, and pgweb from a single interface
- No terminal required
