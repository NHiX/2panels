package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type PathBar struct {
	*tview.Flex
	LeftPath  *tview.TextView
	RightPath *tview.TextView
}

func NewPathBar() *PathBar {
	leftTv := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(tview.AlignLeft)

	leftTv.SetBackgroundColor(tcell.ColorBlack)
	leftTv.SetTextColor(tcell.ColorWhite)

	rightTv := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(tview.AlignLeft)

	rightTv.SetBackgroundColor(tcell.ColorBlack)
	rightTv.SetTextColor(tcell.ColorWhite)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftTv, 0, 1, false).
		AddItem(rightTv, 0, 1, false)

	return &PathBar{
		Flex:      flex,
		LeftPath:  leftTv,
		RightPath: rightTv,
	}
}

func (pb *PathBar) UpdatePaths(leftPath, rightPath string) {
	pb.LeftPath.SetText(fmt.Sprintf(" [yellow]%s[white]", leftPath))
	pb.RightPath.SetText(fmt.Sprintf(" [yellow]%s[white]", rightPath))
}
