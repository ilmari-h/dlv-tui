package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/go-delve/delve/service/api"
	"github.com/rivo/tview"
)

type GoroutinePage struct {
	renderedGoroutines []*api.Goroutine
	commandHandler     *CommandHandler
	listView           *tview.List
	widget             *tview.Frame
}

func NewGoroutinePage() *GoroutinePage {
	listView := tview.NewList()
	listView.SetBackgroundColor(tcell.ColorDefault)
	listView.ShowSecondaryText(false)

	selectedStyle := tcell.StyleDefault.
		Foreground(iToColorTcell(gConfig.Colors.LineFg)).
		Background(iToColorTcell(gConfig.Colors.ListSelectedBg)).
		Attributes(tcell.AttrBold)

	listView.SetSelectedStyle(selectedStyle)

	listView.SetInputCapture(listInputCaptureC)

	pageFrame := tview.NewFrame(listView).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText("[::b]Goroutines:", true, tview.AlignLeft, iToColorTcell(gConfig.Colors.HeaderFg))
	pageFrame.SetBackgroundColor(tcell.ColorDefault)
	gp := GoroutinePage{
		listView: listView,
		widget:   pageFrame,
	}
	return &gp
}

func (page *GoroutinePage) SetCommandHandler(ch *CommandHandler) {
	page.commandHandler = ch
}

func (page *GoroutinePage) RenderGoroutines(grs []*api.Goroutine, currId int) {

	// Filter out external goroutines.
	projectGrs := []*api.Goroutine{}
	for _, gor := range grs {
		if strings.HasPrefix(gor.CurrentLoc.File, page.commandHandler.view.navState.ProjectPath) ||
			strings.HasPrefix(gor.GoStatementLoc.File, page.commandHandler.view.navState.ProjectPath) ||
			strings.HasPrefix(gor.StartLoc.File, page.commandHandler.view.navState.ProjectPath) {
			projectGrs = append(projectGrs, gor)
		}
	}

	page.listView.Clear()
	selectedI := 0
	page.renderedGoroutines = grs
	for i, gor := range projectGrs {
		label := fmt.Sprintf("  [%s]%d.[%s] %s[%s]:%d",
			iToColorS(gConfig.Colors.VarTypeFg),
			gor.ID,
			iToColorS(gConfig.Colors.VarNameFg),
			gor.CurrentLoc.File,
			iToColorS(gConfig.Colors.VarValueFg),
			gor.CurrentLoc.Line,
		)
		if gor.ID == currId {
			selectedI = i
			label = fmt.Sprintf("> [%s::b]%d. [%s]%s[%s]:%d",
				iToColorS(gConfig.Colors.VarTypeFg),
				gor.ID,
				iToColorS(gConfig.Colors.VarNameFg),
				gor.CurrentLoc.File,
				iToColorS(gConfig.Colors.VarValueFg),
				gor.CurrentLoc.Line,
			)
		}
		page.listView.AddItem(
			label,
			"",
			rune(48+i),
			nil).
			SetSelectedFunc(func(i int, s1, s2 string, r rune) {
				page.commandHandler.RunCommand(&SwitchGoroutines{
					Id: page.renderedGoroutines[i].ID,
				})
			})
	}
	page.listView.SetCurrentItem(selectedI)
}

func (sp *GoroutinePage) GetWidget() tview.Primitive {
	return sp.widget
}

func (sp *GoroutinePage) GetName() string {
	return "goroutines"
}

func (sp *GoroutinePage) HandleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	if keyPressed(event, gConfig.Keys.LineDown) {
		sp.listView.SetCurrentItem(sp.listView.GetCurrentItem() + 1)
		return nil
	}
	if keyPressed(event, gConfig.Keys.LineUp) {
		if sp.listView.GetCurrentItem() > 0 {
			sp.listView.SetCurrentItem(sp.listView.GetCurrentItem() - 1)
		}
		return nil
	}
	sp.listView.InputHandler()(event, func(p tview.Primitive) {})
	return nil
}
