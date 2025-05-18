package main

import (
	extractfile "DeadlockHelper/ExtractFile"
	gamebanana "DeadlockHelper/Parser"
	"fmt"

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

	loadBtn := widget.NewButton("Загрузить моды", func() {
		mods, err := gamebanana.FetchMods()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		showModsWindow(a, w, mods, rootInput.Text)
	})

	w.SetContent(container.NewVBox(rootInput, loadBtn))
	w.ShowAndRun()
}

// Открывает новое окно со списком модов
func showModsWindow(a fyne.App, parent fyne.Window, mods []gamebanana.Mod, saveDir string) {
	modsWindow := a.NewWindow("Доступные моды")
	modsWindow.Resize(fyne.NewSize(500, 400))

	list := widget.NewList(
		func() int { return len(mods) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("…"),
				widget.NewButton("Скачать", nil),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			mod := mods[i]
			cont := o.(*fyne.Container)
			cont.Objects[0].(*widget.Label).SetText(mod.Name)
			btn := cont.Objects[1].(*widget.Button)
			btn.OnTapped = func() {
				downloadMod(mod, saveDir, modsWindow)
			}
		},
	)

	modsWindow.SetContent(list)
	modsWindow.Show()
}

// downloadMod запускает загрузку в горутине и показывает диалоги Fyne
// downloadMod запускает загрузку в горутине и показывает диалоги Fyne
func downloadMod(mod gamebanana.Mod, dir string, parent fyne.Window) {
	if dir == "" {
		dialog.ShowError(fmt.Errorf("укажите путь до папки Deadlock"), parent)
		return
	}

	progress := dialog.NewProgressInfinite("Скачивание", fmt.Sprintf("Мод: %s", mod.Name), parent)
	progress.Show()

	go func() {
		outPath, err := gamebanana.DownloadModToDir(mod.ID, dir)

		if err != nil {
			// Показать ошибку через main-горуутину (канал)
			showErrorOnMain(parent, progress, fmt.Errorf("не удалось скачать: %w", err))
			return
		}

		err = extractfile.ExtractAndInstallVPK(outPath, dir)
		if err != nil {
			showErrorOnMain(parent, progress, fmt.Errorf("не удалось установить мод: %w", err))
			return
		}

		showInfoOnMain(parent, progress, fmt.Sprintf("Мод %s установлен", mod.Name))
	}()
}

// Функция для показа ошибки из горутины
func showErrorOnMain(parent fyne.Window, progress dialog.Dialog, err error) {
	go func() {
		// Хак: ждем 1 тик, потом показываем
		progress.Hide()
		dialog.ShowError(err, parent)
	}()
}

// Функция для показа информации
func showInfoOnMain(parent fyne.Window, progress dialog.Dialog, msg string) {
	go func() {
		progress.Hide()
		dialog.ShowInformation("Готово", msg, parent)
	}()
}
