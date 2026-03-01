package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) OpenFileViewer(filePath string) {
	// Try to read up to a reasonable limit (e.g. 1MB) to prevent large files from freezing the UI
	content, err := os.ReadFile(filePath)
	if err != nil {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Could not read file: %s", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.Pages.RemovePage("error")
			})
		a.Pages.AddPage("error", modal, true, true)
		return
	}

	// Basic safety check: don't load huge files
	if len(content) > 1024*1024 {
		content = append(content[:1024*1024], []byte("\n... [content truncated: file too large]")...)
	}

	// Check if binary
	if isBinary(content) {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("%s seems to be a binary file. Open with external editor?", filepath.Base(filePath))).
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.Pages.RemovePage("modal")
				if buttonLabel == "Yes" {
					a.ActionOpenWith(a.getActivePane())
				}
			})
		a.Pages.AddPage("modal", modal, true, true)
		return
	}

	viewer := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetScrollable(true).
		SetText(string(content))

	viewer.SetTitle(fmt.Sprintf(" Viewer: %s ", filePath)).SetBorder(true)
	viewer.SetBackgroundColor(tcell.ColorBlack)
	viewer.SetTextColor(tcell.ColorSilver)

	var targetPane *tview.Table
	var sourcePane *tview.Table
	var leftToRight bool

	focus := a.TviewApp.GetFocus()
	if focus == a.LeftPane.Table {
		targetPane = a.RightPane.Table
		sourcePane = a.LeftPane.Table
		leftToRight = true
	} else if focus == a.RightPane.Table {
		targetPane = a.LeftPane.Table
		sourcePane = a.RightPane.Table
		leftToRight = false
	} else {
		// Default case
		if a.FocusLeft {
			targetPane = a.RightPane.Table
			sourcePane = a.LeftPane.Table
			leftToRight = true
		} else {
			targetPane = a.LeftPane.Table
			sourcePane = a.RightPane.Table
			leftToRight = false
		}
	}

	// Define close function
	closeViewer := func() {
		a.InnerFlex.RemoveItem(viewer)

		// Add the original pane back in the correct order
		if leftToRight { // Target is Right
			a.InnerFlex.AddItem(targetPane, 0, 1, false)
		} else { // Target is Left (need to add it before right pane)
			// Remove right pane to reorder
			a.InnerFlex.RemoveItem(sourcePane)
			a.InnerFlex.AddItem(targetPane, 0, 1, false) // Add left first
			a.InnerFlex.AddItem(sourcePane, 0, 1, true)  // Add right after
		}

		a.TviewApp.SetFocus(sourcePane)
	}

	viewer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyF10 {
			closeViewer()
			return nil
		}
		return event
	})

	// Swap layout
	a.InnerFlex.RemoveItem(targetPane)

	if leftToRight { // Replacing Right
		a.InnerFlex.AddItem(viewer, 0, 1, false)
	} else { // Replacing Left
		// We removed Left, Right is still there. We need Viewer then Right.
		// Remove Right, add Viewer, add Right.
		a.InnerFlex.RemoveItem(sourcePane)
		a.InnerFlex.AddItem(viewer, 0, 1, false)
		a.InnerFlex.AddItem(sourcePane, 0, 1, true)
	}

	a.TviewApp.SetFocus(viewer)
}

func isBinary(data []byte) bool {
	limit := 1024
	if len(data) < limit {
		limit = len(data)
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
