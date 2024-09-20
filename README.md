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
rofi -show robuku
```

### Notes

#### Searching
Tags and URLs are used as metadata for search but are not displayed unless the
bookmark has no title. In that case, the URL is displayed instead of the title.

#### Broken Message Box
If the message box is not resizing to the text, go to your rofi config and remove
the `height` property from `window`. Instead, set the `lines` property
(number of entries listed in rofi) of `listview` to achieve the desired height.

#### Hotkeys (Alt+1, etc.) Not Working
If hotkeys are not working in rofi, check the following properties in the rofi config:
`kb-custom-1`, `kb-custom-2`, and `kb-custom-3`. If they are not set to their default values,
the hotkeys listed in robuku will be incorrect.

## Links

- https://github.com/jarun/buku
- https://github.com/carnager/buku_run (similar project written in bash)
- https://github.com/davatorium/rofi
- https://github.com/lbonn/rofi (wayland fork)
- https://github.com/davatorium/rofi/blob/next/doc/rofi-script.5.markdown

## License

This project is licensed under the MIT License. See the LICENSE file for details.
