package ui

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) ActionSearch(p *Pane) {
	input := tview.NewInputField().
		SetLabel("Search for:").
		SetFieldWidth(40)

	results := tview.NewList().ShowSecondaryText(false)
	results.SetBorder(true).SetTitle("Results")

	closeSearch := func() {
		a.Pages.RemovePage("search_modal")
		a.TviewApp.SetFocus(p.Table)
	}

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			query := input.GetText()
			if query == "" {
				return
			}
			results.Clear()
			results.SetTitle(fmt.Sprintf("Results for: %s", query))

			go func() {
				filepath.WalkDir(p.Path, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					if strings.Contains(strings.ToLower(d.Name()), strings.ToLower(query)) {
						a.TviewApp.QueueUpdateDraw(func() {
							rel, _ := filepath.Rel(p.Path, path)
							results.AddItem(rel, path, 0, func() {
								// On select, navigate to the directory or open the file
								if d.IsDir() {
									p.NavigateTo(path)
								} else {
									p.NavigateTo(filepath.Dir(path))
									// Optional: select the file in the table?
									// For now just navigate to the dir.
								}
								closeSearch()
							})
						})
					}
					return nil
				})
			}()
		} else if key == tcell.KeyEscape {
			closeSearch()
		}
	})

	results.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeSearch()
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			a.TviewApp.SetFocus(input)
			return nil
		}
		return event
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			a.TviewApp.SetFocus(results)
			return nil
		}
		return event
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(input, 3, 1, true).
		AddItem(results, 0, 1, false)

	modalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(flex, 20, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("search_modal", modalFlex, true, true)
	a.TviewApp.SetFocus(input)
}
