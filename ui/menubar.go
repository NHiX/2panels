package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewMenuBar(a *App) *tview.TextView {
	bar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetTextAlign(tview.AlignLeft)

	bar.SetBackgroundColor(tcell.ColorSilver)
	bar.SetTextColor(tcell.ColorBlack)

	bar.SetText(`["left"] [red]L[black]eft [""]  ["right"] [red]R[black]ight [""]  ["file"] [red]F[black]ile [""]  ["archive"] [red]A[black]rchive [""]`)

	bar.SetHighlightedFunc(func(added, removed []string, remaining []string) {
		if len(added) == 0 {
			return
		}
		menuID := added[0]
		a.OpenMenu(menuID)
	})

	return bar
}

func (a *App) OpenMenu(menuID string) {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	list.SetBorder(true).SetTitle(menuID)
	list.SetBackgroundColor(tcell.ColorBlack)

	closeMenu := func() {
		a.Pages.RemovePage("menu")
		a.MenuBar.Highlight()         // Clear highlight
		a.TviewApp.SetFocus(a.Layout) // Restore focus to main layout
	}

	var targetPane *Pane
	switch menuID {
	case "left":
		targetPane = a.LeftPane
	case "right":
		targetPane = a.RightPane
	default:
		// Determine active pane based on focus
		focus := a.TviewApp.GetFocus()
		if focus == a.LeftPane.Table {
			targetPane = a.LeftPane
		} else if focus == a.RightPane.Table {
			targetPane = a.RightPane
		} else {
			if a.FocusLeft {
				targetPane = a.LeftPane
			} else {
				targetPane = a.RightPane
			}
		}
	}

	switch menuID {
	case "left", "right":
		list.AddItem("Sort by Name", "", 'n', func() {
			targetPane.SortBy = "name"
			targetPane.Refresh()
			closeMenu()
		}).
			AddItem("Sort by Size", "", 's', func() {
				targetPane.SortBy = "size"
				targetPane.Refresh()
				closeMenu()
			}).
			AddItem("Sort by Modified", "", 'm', func() {
				targetPane.SortBy = "modified"
				targetPane.Refresh()
				closeMenu()
			}).
			AddItem("Toggle Hidden", "", 'h', func() {
				targetPane.ShowHidden = !targetPane.ShowHidden
				targetPane.Refresh()
				closeMenu()
			})
	case "file":
		list.AddItem("New File", "", 'f', func() {
			closeMenu()
			a.ActionCreateFile(targetPane)
		}).
			AddItem("New Directory", "", 'd', func() {
				closeMenu()
				a.ActionCreateDirectory(targetPane)
			}).
			AddItem("Quit", "", 'q', func() { a.TviewApp.Stop() })
	case "archive":
		list.AddItem("Compress", "", 'c', func() {
			closeMenu()
			a.ActionCompress(targetPane)
		}).
			AddItem("Extract", "", 'x', func() {
				closeMenu()
				a.ActionExtract(targetPane)
			})
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeMenu()
			return nil
		}
		return event
	})

	// Position the menu
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 1, 1, false). // Push below menu bar
			AddItem(list, 10, 1, true).
			AddItem(nil, 0, 1, false), 20, 1, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("menu", flex, true, true)
	a.TviewApp.SetFocus(list)

	// Ensure the correct region is highlighted
	a.MenuBar.Highlight(menuID)
}
