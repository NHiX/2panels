package ui

import (
	"os"

	"github.com/rivo/tview"
)

type App struct {
	TviewApp  *tview.Application
	LeftPane  *Pane
	RightPane *Pane
	Layout    *tview.Flex
	Pages     *tview.Pages
	StatusBar *tview.TextView
	MenuBar   *tview.TextView
	PathBar   *PathBar
	InnerFlex *tview.Flex
	FocusLeft bool
}

func NewApp() *App {
	cwd, _ := os.Getwd()

	app := &App{
		TviewApp:  tview.NewApplication(),
		LeftPane:  NewPane(cwd),
		RightPane: NewPane(cwd),
		FocusLeft: true,
		Pages:     tview.NewPages(),
	}

	// Helper to avoid circular initialization issues if NewStatusBar accesses app
	app.StatusBar = NewStatusBar(app)
	app.MenuBar = NewMenuBar(app)
	app.PathBar = NewPathBar()

	updatePathBar := func(string) {
		app.PathBar.UpdatePaths(app.LeftPane.Path, app.RightPane.Path)
	}
	app.LeftPane.OnPathChange = updatePathBar
	app.RightPane.OnPathChange = updatePathBar

	app.PathBar.UpdatePaths(app.LeftPane.Path, app.RightPane.Path)

	app.InnerFlex = tview.NewFlex().
		AddItem(app.LeftPane.Table, 0, 1, true).
		AddItem(app.RightPane.Table, 0, 1, false)

	app.Layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.MenuBar, 1, 1, false).
		AddItem(app.InnerFlex, 0, 1, true).
		AddItem(app.PathBar, 1, 1, false).
		AddItem(app.StatusBar, 1, 1, false)

	// Setup mouse/enter navigation
	app.LeftPane.SetSelectedFunc(func(row, column int) {
		app.Navigate(app.LeftPane)
	})
	app.RightPane.SetSelectedFunc(func(row, column int) {
		app.Navigate(app.RightPane)
	})

	app.Pages.AddPage("main", app.Layout, true, true)

	app.setupInputCapture()

	return app
}

func (a *App) Run() error {
	return a.TviewApp.SetRoot(a.Pages, true).EnableMouse(true).Run()
}

func (a *App) SwitchFocus() {
	a.FocusLeft = !a.FocusLeft
	if a.FocusLeft {
		a.TviewApp.SetFocus(a.LeftPane.Table)
	} else {
		a.TviewApp.SetFocus(a.RightPane.Table)
	}
}
