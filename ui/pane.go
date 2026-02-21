package ui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Pane represents a file list panel
type Pane struct {
	*tview.Table
	Path         string
	Files        []fs.DirEntry
	OnSelect     func(path string)
	OnPathChange func(path string)
	SortBy       string // "name" (default), "size", "modified"
	ShowHidden   bool
}

// NewPane creates a new Pane
func NewPane(initialPath string) *Pane {
	p := &Pane{
		Table: tview.NewTable().
			SetSelectable(true, false).
			SetFixed(1, 1),
		Path:       initialPath,
		SortBy:     "name",
		ShowHidden: false,
	}
	p.SetBorder(true)

	// Explicit mouse handling
	p.Table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftDoubleClick {
			// Logic handled by SetSelectedFunc usually, but we can force it here if needed
			// Returning the event allows SetSelectedFunc to trigger?
			// Or we can return nil to consume it and call OnSelect manually?
			// Let's rely on standard flow first, but arguably this hook is where we can debug.
			// Actually, tview Table handles DoubleClick to trigger SelectedFunc.
			// So we just pass it through.
		}
		return action, event
	})

	p.Refresh()
	return p
}

// Refresh reloads the file list
func (p *Pane) Refresh() {
	p.Table.Clear()

	// Header
	p.Table.SetCell(0, 0, tview.NewTableCell("Name").SetTextColor(tcell.ColorYellow).SetSelectable(false))
	p.Table.SetCell(0, 1, tview.NewTableCell("Size").SetTextColor(tcell.ColorYellow).SetSelectable(false))
	p.Table.SetCell(0, 2, tview.NewTableCell("Modified").SetTextColor(tcell.ColorYellow).SetSelectable(false))

	// Parent directory entry
	p.Table.SetCell(1, 0, tview.NewTableCell("..").SetTextColor(tcell.ColorGreen))
	p.Table.SetCell(1, 1, tview.NewTableCell("DIR").SetTextColor(tcell.ColorDarkCyan))

	files, err := os.ReadDir(p.Path)
	if err != nil {
		p.Table.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("Error: %s", err)).SetTextColor(tcell.ColorRed))
		return
	}

	var filteredFiles []fs.DirEntry
	for _, f := range files {
		if !p.ShowHidden && len(f.Name()) > 0 && f.Name()[0] == '.' {
			continue
		}
		filteredFiles = append(filteredFiles, f)
	}
	p.Files = filteredFiles

	// Sort: Dirs first, then files sorted by criteria
	sort.Slice(p.Files, func(i, j int) bool {
		d1, d2 := p.Files[i].IsDir(), p.Files[j].IsDir()
		if d1 != d2 {
			return d1
		}

		f1, f2 := p.Files[i], p.Files[j]
		if p.SortBy == "size" {
			info1, err1 := f1.Info()
			info2, err2 := f2.Info()
			if err1 == nil && err2 == nil {
				if info1.Size() != info2.Size() {
					return info1.Size() > info2.Size() // Largest first
				}
			}
		} else if p.SortBy == "modified" {
			info1, err1 := f1.Info()
			info2, err2 := f2.Info()
			if err1 == nil && err2 == nil {
				return info1.ModTime().After(info2.ModTime()) // Newest first
			}
		}

		// Fallback to name or default SortBy name
		return f1.Name() < f2.Name()
	})

	row := 2
	for _, file := range p.Files {
		color := tcell.ColorWhite
		size := ""
		if file.IsDir() {
			color = tcell.ColorBlue
			size = "DIR"
		} else {
			info, err := file.Info()
			if err == nil {
				size = fmt.Sprintf("%d", info.Size())
			}
		}

		// Get ModTime
		modTime := ""
		info, err := file.Info()
		if err == nil {
			modTime = info.ModTime().Format(time.RFC822)
		}

		p.Table.SetCell(row, 0, tview.NewTableCell(file.Name()).SetTextColor(color))
		p.Table.SetCell(row, 1, tview.NewTableCell(size).SetAlign(tview.AlignRight))
		p.Table.SetCell(row, 2, tview.NewTableCell(modTime))

		p.Table.GetCell(row, 0).SetReference(file)

		row++
	}

	if p.OnPathChange != nil {
		p.OnPathChange(p.Path)
	}
}

func (p *Pane) GetSelectedPath() string {
	row, _ := p.GetSelection()
	if row <= 1 { // Header or ..
		return filepath.Dir(p.Path)
	}

	cell := p.GetCell(row, 0)
	ref := cell.GetReference()
	if ref == nil {
		return ""
	}

	entry, ok := ref.(fs.DirEntry)
	if !ok {
		return ""
	}

	return filepath.Join(p.Path, entry.Name())
}

func (p *Pane) NavigateTo(dir string) {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return
	}
	p.Path = dir
	p.Refresh()
	p.Select(1, 0) // Select ".." by default or first element
}
