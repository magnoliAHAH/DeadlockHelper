package main

import (
	updater "DeadlockHelper/SearchPath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Deadlock Helper")
	w.Resize(fyne.NewSize(600, 400))

	input := widget.NewEntry()
	input.SetPlaceHolder("Введите путь до корневой директории Deadlock")

	w.SetContent(
		container.NewVBox(
			input,
			widget.NewButton("Update", func() {
				updater.Update(input.Text)
			}),
		),
	)

	w.ShowAndRun()
}
