# Topotron User Guide

Topotron is a file manager for GNOME, designed for phones running Phosh.
It also works on GNOME desktops and tablets.

You can browse your local files, connect to remote servers over WebDAV, and discover shared folders on your local network — all from one app.

> **Note**: This guide describes touch gestures. On a desktop, "tap" means click, and "long-press" means click and hold.

---

## Table of Contents

- [The Home Screen](#the-home-screen)
- [Browse Files and Folders](#browse-files-and-folders)
- [Open a File](#open-a-file)
- [Select Files](#select-files)
- [Copy and Move Files](#copy-and-move-files)
- [Delete Files](#delete-files)
- [Create a New Folder](#create-a-new-folder)
- [Rename a File or Folder](#rename-a-file-or-folder)
- [Search for Files](#search-for-files)
- [Sort Files](#sort-files)
- [View File Properties](#view-file-properties)
- [Pin a Folder to the Home Screen](#pin-a-folder-to-the-home-screen)
- [Unpin a Folder](#unpin-a-folder)
- [Change a Pinned Folder's Icon](#change-a-pinned-folders-icon)
- [Reset Pinned Folders](#reset-pinned-folders)
- [Connect to a WebDAV Server](#connect-to-a-webdav-server)
- [Browse Network Shares](#browse-network-shares)
- [Show or Hide Hidden Files](#show-or-hide-hidden-files)
- [Keyboard Shortcuts](#keyboard-shortcuts)

---

## The Home Screen

When you open Topotron, you see the home screen.
It has up to three sections:

- **Local** — Pinned folders from your device. By default, this includes Home, Documents, Downloads, Music, Pictures, Videos, and Root.
- **Network** — Shared folders from other devices on your local network. These appear automatically when other devices are sharing files. If no shares are found, this section is hidden.
- **WebDAV** — Servers you have added manually. If you haven't added any, this section is hidden.

Tap any item to start browsing its contents.

The menu button (three horizontal lines) in the top-right corner gives you access to settings and the About dialog.

---

## Browse Files and Folders

1. On the home screen, tap a folder to open it.
2. Inside a folder, you see a list of files and subfolders. Each item shows its name and an icon. Files also show their size.
3. Tap a subfolder to go into it.
4. Tap the back arrow in the top-left corner to go back to the parent folder.

You can keep navigating deeper into subfolders.
The header bar shows the name of the current folder.
To return to the home screen, tap the back arrow until you reach it.

---

## Open a File

Tap a file to open it with your system's default application.
For example, tapping a photo opens it in your image viewer, and tapping a document opens it in your document reader.

---

## Select Files

To perform actions on files (copy, move, delete, rename), you first need to select them.

1. **Long-press** a file or folder. A checkmark appears next to it, and an action bar appears at the bottom of the screen.
2. Tap additional files to add them to the selection.
3. To deselect a file, tap it again.
4. To select everything in the current folder, tap the **Select All** button in the action bar.
5. To cancel the selection and return to normal browsing, tap **Cancel**.

---

## Copy and Move Files

1. [Select the files](#select-files) you want to copy or move.
2. In the action bar at the bottom, tap **Copy** or **Cut**.
   - **Copy** duplicates the files to a new location.
   - **Cut** moves the files to a new location (they are removed from the original folder).
3. A notification appears confirming your choice.
4. Navigate to the destination folder.
5. Tap the **Paste** button in the top-right corner of the header bar.
6. A notification appears when the operation is complete.

The paste button only appears when you have files ready to paste.

---

## Delete Files

1. [Select the files](#select-files) you want to delete.
2. In the action bar at the bottom, tap the **Delete** button (trash icon).
3. The files are deleted and the folder view refreshes.

---

## Create a New Folder

1. While browsing a folder, tap the **New Folder** button (folder with a plus icon) in the top-right corner of the header bar.
2. A new folder named "New Folder" is created. If that name already exists, it becomes "New Folder (2)", "New Folder (3)", and so on.
3. The folder view refreshes to show the new folder.

To give the folder a better name, see [Rename a File or Folder](#rename-a-file-or-folder).

---

## Rename a File or Folder

1. [Select](#select-files) the file or folder you want to rename. You can only rename one item at a time.
2. In the action bar at the bottom, tap the **Rename** button (pencil icon).
3. A dialog appears with the current name. Type the new name.
4. Tap **Rename** to confirm, or **Cancel** to keep the original name.

---

## Search for Files

1. While browsing a folder, tap the **Search** button (magnifying glass icon) in the header bar.
2. A search bar appears. Type part of the file name you are looking for.
3. The file list updates to show only files whose names match your search. The search is not case-sensitive.
4. To clear the search and show all files again, delete the search text or tap the Search button again to close the search bar.

The search filters the current folder only.
It does not search inside subfolders.

---

## Sort Files

1. While browsing a folder, tap the **Sort** button (ascending lines icon) in the header bar.
2. A menu appears with sort options:
   - **Name (A to Z)** — alphabetical order (default)
   - **Name (Z to A)** — reverse alphabetical order
   - **Date (oldest)** — oldest files first
   - **Date (newest)** — newest files first
   - **Size (smallest)** — smallest files first
   - **Size (largest)** — largest files first
3. Tap an option to apply it. A checkmark shows the current sort order.

Folders always appear before files, regardless of the sort order.
Your sort preference is saved and applied to all folders.

---

## View File Properties

1. [Select](#select-files) a single file or folder.
2. In the action bar at the bottom, tap the **Properties** button (info icon).
3. A properties page appears showing:
   - **Type** — file or folder
   - **Location** — the full path
   - **Size** — for files, the size in a readable format with the exact byte count; for folders, the total size is calculated in the background
   - **Modified** — the date and time the item was last changed

Tap the back arrow to return to the file browser.

---

## Pin a Folder to the Home Screen

Pinning adds a folder to the "Local" section of the home screen for quick access.

1. Navigate to the folder you want to pin.
2. [Select](#select-files) it by long-pressing.
3. In the action bar at the bottom, tap the **Pin** button (pin icon).
4. A notification confirms the folder has been pinned.
5. The folder now appears on the home screen.

You can pin any folder, including folders on WebDAV servers.

---

## Unpin a Folder

1. On the home screen, **long-press** the pinned folder you want to remove.
2. An action bar appears at the bottom with the folder's name.
3. Tap **Unpin** to remove it from the home screen.

Unpinning a folder does not delete it.
It only removes the shortcut from the home screen.

---

## Change a Pinned Folder's Icon

1. On the home screen, **long-press** the pinned folder whose icon you want to change.
2. An action bar appears at the bottom. Tap the **Change Icon** button (appearance icon).
3. An icon picker appears with a grid of available icons (folder, home, documents, downloads, music, pictures, videos, and more).
4. Tap the icon you want. The current icon is highlighted.
5. The home screen updates to show the new icon.

---

## Reset Pinned Folders

If you want to restore the default set of pinned folders:

1. On the home screen, tap the **menu button** (three horizontal lines) in the top-right corner.
2. Tap **Reset Pinned Folders**.
3. The pinned folders are reset to the defaults: Home, Documents, Downloads, Music, Pictures, Videos, and Root.

This removes any custom pins you have added and restores any defaults you have removed.

---

## Connect to a WebDAV Server

WebDAV lets you browse files on a remote server as if they were on your device. Many services support WebDAV, including Nextcloud, ownCloud, and most NAS devices.

1. On the home screen, tap **Add WebDAV Server**.
2. Fill in the connection details:
   - **Name** — a friendly name for this server (for example, "My Nextcloud")
   - **URL** — the WebDAV URL (for example, `https://cloud.example.com/remote.php/dav/files/username/`)
   - **Username** — your username on the server
   - **Password** — your password
3. Tap **Add**.
4. The server appears in the "WebDAV" section of the home screen.
5. Tap it to connect and browse.

All file operations (copy, move, delete, rename, new folder) work on WebDAV servers just like they do on local files.

---

## Browse Network Shares

Topotron automatically discovers shared folders from other devices on your local network. This works with:

- Other GNOME desktops that have file sharing turned on
- NAS devices (Synology, QNAP, and others)
- Any device advertising a WebDAV service

Discovered shares appear in the **Network** section of the home screen.
You don't need to configure anything — they show up automatically when devices are available and disappear when they go offline.

Tap a network share to browse its contents.
You can copy files between network shares and your local folders.

---

## Show or Hide Hidden Files

Files and folders whose names start with a dot (like `.bashrc` or `.config`) are hidden by default.

1. On the home screen, tap the **menu button** in the top-right corner.
2. Toggle **Show Hidden Files** on or off.

This setting is saved and applies to all folders.

---

## Keyboard Shortcuts

When using Topotron on a desktop with a keyboard:

| Shortcut   | Action        |
|------------|---------------|
| **Ctrl+Q** | Quit Topotron |
