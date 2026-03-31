# topotron

**topotron** is a file manager for GNOME, GTK 4, and Libadwaita optimized for a mobile user-experience.

It supports local filesystems, remote filesystems via WebDAV, and automatically discovers shared folders on your local network.

## Other Files

* The **user guide** for **topotron** is at: [GUIDE.md](GUIDE.md)
* THe **developer guide** for **topotron** is at: [HACKING.md](HACKING.md)

## Build

### Requirements

- Go >= 1.21
- GTK 4 >= 4.10 (development headers)
- Libadwaita >= 1.4 (development headers)
- GLib 2.0 (development headers)

On Fedora:
```bash
sudo dnf install gtk4-devel libadwaita-devel glib2-devel
```

On Debian/Ubuntu:
```bash
sudo apt install libgtk-4-dev libadwaita-1-dev libglib2.0-dev
```

### Development Build

```bash
go build
```

Or with vendored dependencies:
```bash
go build -mod=vendor
```

### Run

```bash
./topotron
```

### Flatpak

```bash
flatpak install --user org.gnome.{Platform,Sdk}//47
flatpak install --user org.freedesktop.Sdk.Extension.golang//24.08
flatpak-builder --user --force-clean --install build build-aux/flatpak/link.reiver.topotron.json
flatpak run link.reiver.topotron
```

## Author

Software **topotron** was written by [Charles Iliya Krempeaux](http://reiver.link)
