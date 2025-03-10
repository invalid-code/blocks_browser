package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Blocks Browser")

	searchBar := widget.NewEntry()
	searchBar.PlaceHolder = "Search the web"
	refreshBtn := widget.NewButtonWithIcon("", func() {})
	toolBar := container.NewVBox(refreshBtn, searchBar)
	w.SetContent(toolBar)
	w.ShowAndRun()
}
