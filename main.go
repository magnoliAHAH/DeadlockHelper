package main

import (
	extractfile "DeadlockHelper/ExtractFile"
	gamebanana "DeadlockHelper/Parser"
	updater "DeadlockHelper/SearchPath"
	installlog "DeadlockHelper/installedmods"
	"fmt"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Deadlock Helper")
	w.Resize(fyne.NewSize(600, 400))

	rootInput := widget.NewEntry()
	rootInput.SetPlaceHolder("Введите путь до папки Deadlock")

	loadBtn := widget.NewButton("Загрузить моды", func() {
		// 1. Создаём бесконечный прогресс-бар
		progressBar := widget.NewProgressBarInfinite()

		// 2. Упаковываем его в кастомный диалог без кнопок и показываем
		loadingDialog := dialog.NewCustomWithoutButtons(
			"Загрузка модов", progressBar, w,
		)
		loadingDialog.Show()

		// 3. Запускаем загрузку в фоне
		go func() {
			mods, err := gamebanana.FetchMods(1)

			// 4. Безопасно обновляем UI из горутины
			fyne.Do(func() {
				// останавливаем анимацию и скрываем диалог
				progressBar.Stop()
				loadingDialog.Hide()

				// обрабатываем результат
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				showModsWindow(a, w, mods, rootInput.Text)
			})
		}()
	})

	updateBtn := widget.NewButton("Обновить путь", func() {
		err := updater.Update(rootInput.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Готово", "Файл gameinfo.gi обновлён", w)
	})

	w.SetContent(container.NewVBox(
		rootInput,
		container.NewHBox(loadBtn, updateBtn),
	))
	w.ShowAndRun()
}

func showModsWindow(a fyne.App, parent fyne.Window, initialMods []gamebanana.Mod, saveDir string) {
	modsWindow := a.NewWindow("Доступные моды")
	modsWindow.Resize(fyne.NewSize(800, 600))

	currentPage := 1
	var items []fyne.CanvasObject

	grid := container.NewGridWithColumns(3)
	scroll := container.NewVScroll(grid)

	loadMoreBtn := widget.NewButton("Загрузить ещё", func() {
		// Создаем бесконечный индикатор прогресса
		progressBar := widget.NewProgressBarInfinite()
		progressBar.Start()

		// Создаем пользовательский диалог без кнопок
		customDialog := dialog.NewCustomWithoutButtons("Загрузка", progressBar, modsWindow)
		customDialog.Show()

		// Запускаем длительную операцию в отдельной горутине
		go func() {
			currentPage++
			newMods, err := gamebanana.FetchMods(currentPage)

			// Обновляем интерфейс в главном потоке
			fyne.Do(func() {
				progressBar.Stop()
				customDialog.Hide()
				if err != nil {
					dialog.ShowError(err, modsWindow)
				} else {
					addModsToGrid(newMods, &items, grid, saveDir, modsWindow)
				}
			})
		}()
	})

	addModsToGrid(initialMods, &items, grid, saveDir, modsWindow)

	content := container.NewBorder(nil, loadMoreBtn, nil, nil, scroll)
	modsWindow.SetContent(content)
	modsWindow.Show()
}

func addModsToGrid(mods []gamebanana.Mod, items *[]fyne.CanvasObject, grid *fyne.Container, saveDir string, parent fyne.Window) {
	for _, mod := range mods {
		modCopy := mod

		var img fyne.CanvasObject = widget.NewLabel("Загрузка изображения...")
		if mod.ImageURL() != "" {
			if uri, err := url.Parse(mod.ImageURL()); err == nil {
				image := canvas.NewImageFromURI(storage.NewURI(uri.String()))
				image.FillMode = canvas.ImageFillContain
				image.SetMinSize(fyne.NewSize(150, 150))
				img = image
			}
		}

		card := container.NewVBox(
			img,
			widget.NewLabel(mod.Name),
			widget.NewButton("Скачать", func() {
				downloadMod(modCopy, saveDir, parent)
			}),
		)
		cardContainer := container.NewBorder(nil, nil, nil, nil, card)
		*items = append(*items, cardContainer)
		grid.Add(cardContainer)
	}
}

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
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("не удалось скачать: %w", err), parent)
			})
			return
		}

		modPath, err := extractfile.ExtractAndInstallVPK(outPath, dir)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("не удалось установить мод: %w", err), parent)
			})
			return
		}

		// ✅ Сохраняем информацию об установленном моде
		_ = installlog.SaveInstalledMod(installlog.InstalledMod{
			ID:        mod.ID,
			Name:      mod.Name,
			ImageURL:  mod.ImageURL(),
			Path:      modPath, // путь до скачанного архива (или измените на финальный путь, если нужно)
			Installed: time.Now(),
		}, dir)

		fyne.Do(func() {
			progress.Hide()
			dialog.ShowInformation("Успех", fmt.Sprintf("Мод %s установлен успешно", mod.Name), parent)
		})
	}()
}
