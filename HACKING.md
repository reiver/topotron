# Topotron Developer Guide

## What is Topotron

Topotron is a GTK 4 / Libadwaita file manager for GNOME, optimized for mobile (Phosh).
It supports local filesystems, WebDAV remote filesystems, and automatic mDNS/Avahi network share discovery.

App ID: `link.reiver.topotron`

## Build & Run

```bash
go build -v # buikd

./topotron  # run
```

System dependencies (Fedora): `gtk4-devel libadwaita-devel glib2-devel`

Meson is used for Flatpak/install builds but `go build` is sufficient for development.

## Architecture

The codebase follows a **cfg/lib/srv** layered architecture:

- **`cfg/`** — Application constants (app ID, version, app path) and configuration (log level via `LOG_LEVEL` env var).
- **`gui/`** — All GTK UI code. Package alias is `gui`. Entry point is `gui.Run()` which creates the Adwaita app and main window. The `Window` struct owns navigation, clipboard state, and wires up all page callbacks.
- **`lib/`** — Pure library code with no dependencies on GTK or services:
  - `lib/backend/` — `FileBackend` interface abstracting file operations, with `LocalBackend` (OS filesystem) and `WebDAVBackend` implementations.
  - `lib/fileinfo/` — File metadata type used across backends.
  - `lib/format/` — Display formatting (file sizes).
  - `lib/icon/` — Icon resolution.
  - `lib/place/` — Place/location type for the home screen.
- **`srv/`** — Service layer (stateful, side-effectful):
  - `srv/settings/` — Persistent user preferences stored as JSON at `~/.config/topotron/settings.json`.
  - `srv/discover/` — mDNS/Avahi service discovery for `_webdav._tcp` shares via D-Bus.
  - `srv/op/` — File operation orchestration (copy/move with progress callbacks).
  - `srv/log/` — Structured logging via `codeberg.org/reiver/go-log`.

## Key Patterns

- GTK bindings are from `github.com/diamondburned/gotk4` and `gotk4-adwaita`.
- Background work dispatches back to the GTK main thread via `glib.IdleAdd()`.
- Method receivers are named `receiver` (not short abbreviations).
- Nil checks use inverted style: `if nil == err` / `if nil != err`.

## Data & Resources

- `data/` — Desktop entry, GSchema, metainfo, CSS, icons, and GResource XML for Meson/Flatpak packaging.
- `po/` — i18n translation files.
- `build-aux/flatpak/` — Flatpak build manifest.
