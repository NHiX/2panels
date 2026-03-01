package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"2panels/fs"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) ActionCreateFile(p *Pane) {
	a.promptInput("New File Name:", func(name string) {
		if name == "" {
			return
		}
		path := filepath.Join(p.Path, name)
		file, err := os.Create(path)
		if err != nil {
			a.showError(fmt.Sprintf("Failed to create file: %v", err))
			return
		}
		file.Close()
		p.Refresh()
	})
}

func (a *App) ActionCreateDirectory(p *Pane) {
	a.promptInput("New Directory Name:", func(name string) {
		if name == "" {
			return
		}
		path := filepath.Join(p.Path, name)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			a.showError(fmt.Sprintf("Failed to create directory: %v", err))
			return
		}
		p.Refresh()
	})
}

func (a *App) promptInput(title string, onSubmit func(text string)) {
	input := tview.NewInputField().
		SetLabel(title).
		SetFieldWidth(40)

	input.SetDoneFunc(func(key tcell.Key) {
		text := input.GetText()
		a.Pages.RemovePage("input_modal")
		if key == tcell.KeyEnter {
			onSubmit(text)
		}
		a.TviewApp.SetFocus(a.Layout)
	})

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(input, 3, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("input_modal", flex, true, true)
	a.TviewApp.SetFocus(input)
}

func (a *App) showError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.Pages.RemovePage("error_modal")
		})
	a.Pages.AddPage("error_modal", modal, true, true)
}

func (a *App) ActionCompress(p *Pane) {
	srcPath := p.GetSelectedPath()
	if srcPath == "" {
		return
	}

	defaultName := filepath.Base(srcPath) + ".zip"
	a.promptInput(fmt.Sprintf("Archive Name (%s):", defaultName), func(name string) {
		if name == "" {
			name = defaultName
		}

		// Detect extension
		ext := filepath.Ext(name)
		if ext == "" {
			name += ".zip"
			ext = ".zip"
		}

		archiver, err := fs.GetArchiver(ext)
		if err != nil {
			a.showError(fmt.Sprintf("Format not supported: %v", err))
			return
		}

		dstPath := filepath.Join(p.Path, name)
		go func() {
			err := archiver.Compress(srcPath, dstPath)
			a.TviewApp.QueueUpdateDraw(func() {
				if err != nil {
					a.showError(fmt.Sprintf("Compression failed: %v", err))
				} else {
					p.Refresh()
				}
			})
		}()
	})
}

func (a *App) ActionExtract(p *Pane) {
	srcPath := p.GetSelectedPath()
	if srcPath == "" {
		return
	}

	// Simple extension detection
	ext := filepath.Ext(srcPath)
	if strings.HasSuffix(strings.ToLower(srcPath), ".tar.gz") {
		ext = ".tar.gz"
	}

	archiver, err := fs.GetArchiver(ext)
	if err != nil {
		a.showError("Please select a supported archive file to extract.")
		return
	}

	dstDir := p.Path
	a.promptInput(fmt.Sprintf("Extract to (%s):", dstDir), func(target string) {
		if target == "" {
			target = dstDir
		}

		go func() {
			err := archiver.Extract(srcPath, target)
			a.TviewApp.QueueUpdateDraw(func() {
				if err != nil {
					a.showError(fmt.Sprintf("Extraction failed: %v", err))
				} else {
					p.Refresh()
				}
			})
		}()
	})
}
