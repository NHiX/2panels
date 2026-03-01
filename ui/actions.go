package ui

import (
	"2panels/fs"
	"fmt"
	"os"
	"os/exec"
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

func (a *App) ActionDelete() {
	var srcPane *Pane
	focus := a.TviewApp.GetFocus()
	if focus == a.LeftPane.Table {
		srcPane = a.LeftPane
	} else if focus == a.RightPane.Table {
		srcPane = a.RightPane
	} else {
		if a.FocusLeft {
			srcPane = a.LeftPane
		} else {
			srcPane = a.RightPane
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

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete %s? This action cannot be undone.", filepath.Base(srcPath))).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				go func() {
					err := fs.Delete(srcPath)
					a.TviewApp.QueueUpdateDraw(func() {
						if err != nil {
							a.showError(fmt.Sprintf("Error deleting %s: %v", srcPath, err))
						} else {
							srcPane.Refresh()
						}
						a.Pages.RemovePage("delete_modal")
					})
				}()
			} else {
				a.Pages.RemovePage("delete_modal")
			}
		})

	a.Pages.AddPage("delete_modal", modal, true, true)
}
func (a *App) ActionRename() {
	var srcPane *Pane
	focus := a.TviewApp.GetFocus()
	if focus == a.LeftPane.Table {
		srcPane = a.LeftPane
	} else if focus == a.RightPane.Table {
		srcPane = a.RightPane
	} else {
		if a.FocusLeft {
			srcPane = a.LeftPane
		} else {
			srcPane = a.RightPane
		}
	}

	srcPath := srcPane.GetSelectedPath()
	if srcPath == "" {
		return
	}

	oldName := filepath.Base(srcPath)
	a.promptInput(fmt.Sprintf("Rename %s to:", oldName), func(newName string) {
		if newName == "" || newName == oldName {
			return
		}

		dstPath := filepath.Join(filepath.Dir(srcPath), newName)
		err := os.Rename(srcPath, dstPath)
		if err != nil {
			a.showError(fmt.Sprintf("Error renaming: %v", err))
		} else {
			srcPane.Refresh()
		}
	})
}

func (a *App) ActionOpenWith(p *Pane) {
	srcPath := p.GetSelectedPath()
	if srcPath == "" {
		return
	}

	info, err := os.Stat(srcPath)
	if err != nil || info.IsDir() {
		return
	}

	editors := []string{os.Getenv("EDITOR"), "vim", "nano", "vi"}
	var finalEditor string
	for _, ed := range editors {
		if ed == "" {
			continue
		}
		if _, err := exec.LookPath(ed); err == nil {
			finalEditor = ed
			break
		}
	}

	if finalEditor == "" {
		a.showError("No editor found on system. Please set $EDITOR.")
		return
	}

	err = a.SuspendAndRun(finalEditor, []string{srcPath})
	if err != nil {
		a.showError(fmt.Sprintf("Error launching editor: %v", err))
	}
	p.Refresh()
}
