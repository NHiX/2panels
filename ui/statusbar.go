package ui

import (
	"2panels/fs"
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

	formatSize := func(b uint64) string {
		const unit = 1024
		if b < unit {
			return fmt.Sprintf("%d B", b)
		}
		div, exp := uint64(unit), 0
		for n := b / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
	}

	updateText := func() {
		currentTime := time.Now().Format("15:04:05")
		diskInfo := ""
		usage, err := fs.GetDiskUsage(a.LeftPane.Path)
		if err == nil {
			diskInfo = fmt.Sprintf(" | Disk: %s/%s free", formatSize(usage.Free), formatSize(usage.Total))
		}

		text := fmt.Sprintf(`%s%s | ["open"] [yellow]F4[white] Open With [""]  ["copy"] [yellow]F5[white] Copy [""]  ["move"] [yellow]F6[white] Move [""]  ["quit"] [yellow]F10[white] Quit [""]`, currentTime, diskInfo)
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
		case "open":
			a.ActionOpenWith(a.getActivePane())
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
