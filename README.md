# Robuku

**Robuku** is a rofi script written in Go that manages bookmarks from buku.
This script allows you to add, delete, modify, and navigate to bookmarks directly
from the Rofi interface.

## Features

- **Search**: Find bookmarks using rofi's built-in matching features.
- **Goto**:   Open a bookmark in your default web browser.
- **Add**:    Add a new bookmark.
- **Delete**: Remove an existing bookmark.
- **Modify**: Update fields of an existing bookmark.

## Requirements

- A buku SQLite database file (`bookmarks.db`). You can set a custom path with the environment variable `$ROBUKU_DB_PATH`.
- Optionally, `xdg-utils` or you can set a browser with the environment variable `$ROBUKU_BROWSER`.

## Installation

### Build from source
1. Clone the repository:
    ```sh
    git clone --depth 1 https://github.com/vannrr/robuku.git
    cd robuku
    ```

2. Build Robuku:
    - a.
    ```sh
    make install
    ```
    - b.
    or if you don't have `make`:
    ```sh
    go build -ldflags="-w -s" -o robuku main.go
    mkdir -p ~/.config/rofi/scripts/
    cp robuku ~/.config/rofi/scripts/
    ```

### Install from binary
1. Download from release page https://github.com/vannrr/robuku/releases/latest
2. place binary in `~/.config/rofi/scripts/`

## Usage

Run the script with rofi:
```sh
rofi -show robuku -modi robuku
```

**Note** Tags and urls are used as metadata for search but not displayed unless
the bookmark has no title then the url is displayed.

## Links

- https://github.com/jarun/buku
- https://github.com/carnager/buku_run (similar project written in bash)
- https://github.com/davatorium/rofi
- https://github.com/lbonn/rofi (wayland fork)
- https://github.com/davatorium/rofi/blob/next/doc/rofi-script.5.markdown

## License

This project is licensed under the MIT License. See the LICENSE file for details.
