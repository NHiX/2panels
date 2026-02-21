package ui

import (
	"2panels/fs"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rivo/tview"
)

func (a *App) Navigate(p *Pane) {
	selection := p.GetSelectedPath()
	if selection != "" {
		// Check if it's a dir
		info, err := os.Stat(selection)
		if err == nil && info.IsDir() {
			p.NavigateTo(selection)
		} else if err == nil {
			a.OpenFileViewer(selection)
		}
	} else {
		// It is ".."
		p.NavigateTo(filepath.Dir(p.Path))
	}
}

func (a *App) Action(move bool) {
	var srcPane, dstPane *Pane

	// Determine active pane based on focus
	focus := a.TviewApp.GetFocus()
	if focus == a.LeftPane.Table {
		srcPane = a.LeftPane
		dstPane = a.RightPane
	} else if focus == a.RightPane.Table {
		srcPane = a.RightPane
		dstPane = a.LeftPane
	} else {
		// Default to FocusLeft state or LeftPane if logic fails
		if a.FocusLeft {
			srcPane = a.LeftPane
			dstPane = a.RightPane
		} else {
			srcPane = a.RightPane
			dstPane = a.LeftPane
		}
	}

	srcPath := srcPane.GetSelectedPath()
	if srcPath == "" {
		return
	}

	row, _ := srcPane.GetSelection()
	if row <= 1 {
		return
	}

	dstDir := dstPane.Path
	actionName := "Copy"
	if move {
		actionName = "Move"
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("%s %s to %s?", actionName, filepath.Base(srcPath), dstDir)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				go func() {
					var err error
					if move {
						err = fs.Move(srcPath, dstDir)
					} else {
						err = fs.Copy(srcPath, dstDir)
					}

					a.TviewApp.QueueUpdateDraw(func() {
						if err != nil {
							// Show error
							errModal := tview.NewModal().
								SetText(fmt.Sprintf("Error: %s", err)).
								AddButtons([]string{"OK"}).
								SetDoneFunc(func(buttonIndex int, buttonLabel string) {
									a.Pages.RemovePage("error")
								})
							a.Pages.AddPage("error", errModal, true, true)
						} else {
							// Refresh both panes
							srcPane.Refresh()
							dstPane.Refresh()
						}
						a.Pages.RemovePage("modal")
					})
				}()
			} else {
				a.Pages.RemovePage("modal")
			}
		})

	a.Pages.AddPage("modal", modal, true, true)
}
