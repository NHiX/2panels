package ui

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// In a real app, this would be loaded from a config file
var globalBookmarks = []string{"/", os.Getenv("HOME")}

func (a *App) ActionBookmarks(p *Pane) {
	list := tview.NewList().ShowSecondaryText(true).SetHighlightFullLine(true)
	list.SetBorder(true).SetTitle("Bookmarks (Ctrl+B)")

	closeBookmarks := func() {
		a.Pages.RemovePage("bookmarks_modal")
		a.TviewApp.SetFocus(p.Table)
	}

	updateList := func() {
		list.Clear()
		for i, b := range globalBookmarks {
			path := b
			list.AddItem(b, "", rune('1'+i), func() {
				p.NavigateTo(path)
				closeBookmarks()
			})
		}
		list.AddItem("Add Current Path", p.Path, 'a', func() {
			globalBookmarks = append(globalBookmarks, p.Path)
			closeBookmarks()
			a.ActionBookmarks(p) // Refresh
		})
	}

	updateList()

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeBookmarks()
			return nil
		}
		return event
	})

	modalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, 15, 1, true).
			AddItem(nil, 0, 1, false), 50, 1, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("bookmarks_modal", modalFlex, true, true)
	a.TviewApp.SetFocus(list)
}
