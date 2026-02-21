package ui

import (
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

func (a *App) setupInputCapture() {
	a.TviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			a.SwitchFocus()
			return nil
		case tcell.KeyF5:
			a.Action(false) // Copy
			return nil
		case tcell.KeyF6:
			a.Action(true) // Move
			return nil
		case tcell.KeyF10:
			a.TviewApp.Stop()
			return nil
		case tcell.KeyCtrlL:
			a.OpenMenu("left")
			return nil
		case tcell.KeyCtrlR:
			a.OpenMenu("right")
			return nil
		case tcell.KeyCtrlF:
			a.OpenMenu("file")
			return nil
		case tcell.KeyCtrlA:
			a.OpenMenu("archive")
			return nil
		}

		/* switch event.Rune() {
		case 'q':
			a.TviewApp.Stop()
		} */

		return event
	})

	// Pane specific navigation
	handlePaneInput := func(p *Pane, nextPane *Pane) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			/* case tcell.KeyEnter:
			                // Handled by SetSelectedFunc (Enter or Double Click)
							return nil */
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				p.NavigateTo(filepath.Dir(p.Path))
				return nil
			}
			return event
		}
	}

	a.LeftPane.SetInputCapture(handlePaneInput(a.LeftPane, a.RightPane))
	a.RightPane.SetInputCapture(handlePaneInput(a.RightPane, a.LeftPane))
}
