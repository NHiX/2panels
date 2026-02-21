package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewStatusBar(a *App) *tview.TextView {
	bar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetTextAlign(tview.AlignLeft)

	bar.SetBackgroundColor(tcell.ColorDarkBlue)
	bar.SetTextColor(tcell.ColorWhite)

	updateText := func() {
		currentTime := time.Now().Format("15:04:05")
		text := fmt.Sprintf(`%s | ["copy"] [yellow]F5[white] Copy [""]  ["move"] [yellow]F6[white] Move [""]  ["quit"] [yellow]F10[white] Quit [""]`, currentTime)
		bar.SetText(text)
	}

	updateText() // Initial paint

	// Periodic clock update
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			a.TviewApp.QueueUpdateDraw(func() {
				updateText()
			})
		}
	}()

	bar.SetHighlightedFunc(func(added, removed []string, remaining []string) {
		if len(added) == 0 {
			return
		}

		switch added[0] {
		case "copy":
			a.Action(false)
		case "move":
			a.Action(true)
		case "quit":
			a.TviewApp.Stop()
		}

		// Clear highlight so it can be clicked again
		bar.Highlight()
	})

	return bar
}
