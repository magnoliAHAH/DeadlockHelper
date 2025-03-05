package main

import (
	updater "DeadlockHelper/SearchPath"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Deadlock Helper")
	w.Resize(fyne.NewSize(600, 400))

	rootInput := widget.NewEntry()
	rootInput.SetPlaceHolder("Введите путь до папки Deadlock")

	w.SetContent(container.NewVBox(
		rootInput,
		widget.NewButton("Update", func() {
			updater.Update(rootInput.Text)
		}),
		widget.NewButton("Manage Addons", func() {
			openAddonsWindow(a, rootInput.Text) // Передаем текст из rootInput
		}),
	))

	w.ShowAndRun()
}

func openAddonsWindow(a fyne.App, rootPath string) {
	skinsFolder := filepath.Join(rootPath, "game/citadel/addons") // Обновляем путь к папке addons
	fmt.Println(skinsFolder)
	w := a.NewWindow("Manage Addons")
	w.Resize(fyne.NewSize(500, 400))

	getFiles := func() []os.DirEntry {
		files, err := os.ReadDir(skinsFolder)
		if err != nil {
			return []os.DirEntry{}
		}
		return files
	}
	var fileList *widget.List
	fileList = widget.NewList(
		func() int {
			return len(getFiles())
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewButton("Delete", func() {}),
			)
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			files := getFiles()
			if i < len(files) {
				label := obj.(*fyne.Container).Objects[0].(*widget.Label)
				label.SetText(files[i].Name())

				btn := obj.(*fyne.Container).Objects[1].(*widget.Button)
				btn.OnTapped = func() {
					os.Remove(filepath.Join(skinsFolder, files[i].Name())) // Используем обновленный путь
					fileList.Refresh()
				}
			}
		},
	)

	addFileBtn := widget.NewButton("Add File", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				dstPath := filepath.Join(skinsFolder, reader.URI().Name()) // Используем обновленный путь
				data, _ := io.ReadAll(reader)
				os.WriteFile(dstPath, data, 0644)
				fileList.Refresh()
			}
		}, w)
	})

	w.SetContent(container.NewBorder(nil, addFileBtn, nil, nil, fileList))
	w.Show()
}
