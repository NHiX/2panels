# 2panels

2panels is a modern, terminal-based, dual-pane file manager built in Go using the `tview` and `tcell` libraries.

![Overview](https://github.com/NHiX/2panels/blob/master/screenshot.png?raw=true) *(Screenshot coming soon)*

## Features

- **Dual-Pane Interface**: Navigate two different directories simultaneously for easy file operations.
- **Fast Navigation**: Keyboard and mouse support for quick browsing.
- **File Operations**:
  - Copy (`F5`)
  - Move (`F6`)
  - Delete / Rename (Coming soon)
  - Create New File / Directory (`Ctrl+F`)
- **Text File Viewer**: Select any text file to preview its contents instantly in the opposite pane.
- **Sorting & Filtering**:
  - Sort by Name, Size, or Modification Date (`Ctrl+L` or `Ctrl+R`).
  - Toggle visibility of hidden files.
- **Cross-Platform**: Works smoothly on macOS, Linux, and Windows terminals.

## Installation

### Prerequisites
Make sure you have [Go](https://golang.org/dl/) installed (version 1.20 or later recommended).

### Build from source
Clone the repository and build the executable:

```sh
git clone https://github.com/NHiX/2panels.git
cd 2panels
go build
```

Then run the binary:
```sh
./2panels
```

## Keyboard Shortcuts

| Shortcut | Action |
| --- | --- |
| `Tab` | Switch focus between left and right pane |
| `Up` / `Down` | Navigate within a pane |
| `Enter` | Enter directory / Open text file viewer |
| `Backspace` | Go up one directory (`..`) |
| `F5` | Copy selected item to the other pane |
| `F6` | Move selected item to the other pane |
| `F10` | Quit application / Close menus and viewers |
| `Esc` | Close open menus and text viewer |
| `Ctrl+L` | Open Left Pane options menu (Sort, Hidden files) |
| `Ctrl+R` | Open Right Pane options menu (Sort, Hidden files) |
| `Ctrl+F` | Open File menu (New File, New Dir, Quit) |
| `Ctrl+A` | Open Archive menu (Compress, Extract) |

## Tech Stack

- [Go](https://go.dev/)
- [tview](https://github.com/rivo/tview) - Rich interactive widgets for terminal-based UIs
- [tcell](https://github.com/gdamore/tcell) - Cell based view for text terminals

## License

This project is licensed under the MIT License.
