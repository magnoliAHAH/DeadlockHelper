package main

import (
	config "DeadlockHelper/Config"
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
	iconRes, err := fyne.LoadResourceFromPath("DeadlockHelper_icon.ico")
	if err != nil {
		fmt.Println("Ошибка загрузки иконки:", err)
	} else {
		a.SetIcon(iconRes)
	}

	// Загрузка конфига
	cfg, err := config.LoadConfig()
	if err != nil {
		dialog.ShowError(fmt.Errorf("ошибка загрузки конфига: %w", err), w)
	}

	rootInput := widget.NewEntry()
	rootInput.SetPlaceHolder("Введите путь до папки Deadlock")

	statusLabel := widget.NewLabel("")
	if cfg.DeadlockPath == "" {
		statusLabel.SetText("Конфиг не найден, введите директорию Deadlock и Сохранить")
	} else {
		rootInput.SetText(cfg.DeadlockPath)
	}

	savePathBtn := widget.NewButton("Сохранить путь", func() {
		path := rootInput.Text
		if path == "" {
			dialog.ShowError(fmt.Errorf("путь не может быть пустым"), w)
			return
		}
		err := config.SaveConfig(config.Config{DeadlockPath: path})
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка сохранения конфига: %w", err), w)
			return
		}
		statusLabel.SetText("") // убираем сообщение
		dialog.ShowInformation("Успех", "Путь успешно сохранён", w)
	})

	loadBtn := widget.NewButton("Загрузить моды", func() {
		progressBar := widget.NewProgressBarInfinite()
		loadingDialog := dialog.NewCustomWithoutButtons("Загрузка модов", progressBar, w)
		loadingDialog.Show()

		go func() {
			mods, err := gamebanana.FetchMods(1)
			fyne.Do(func() {
				progressBar.Stop()
				loadingDialog.Hide()
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

	installedBtn := widget.NewButton("Установленные моды", func() {
		showInstalledModsWindow(a, w, rootInput.Text)
	})

	w.SetContent(container.NewVBox(
		statusLabel,
		rootInput,
		savePathBtn,
		container.NewHBox(loadBtn, updateBtn, installedBtn),
	))

	w.ShowAndRun()
}

func showInstalledModsWindow(a fyne.App, parent fyne.Window, dir string) {
	if dir == "" {
		dialog.ShowError(fmt.Errorf("укажите путь до папки Deadlock"), parent)
		return
	}

	mods, err := installlog.LoadInstalledMods(dir)
	if err != nil {
		dialog.ShowError(fmt.Errorf("не удалось загрузить установленные моды: %w", err), parent)
		return
	}

	window := a.NewWindow("Установленные моды")
	window.Resize(fyne.NewSize(800, 600))

	grid := container.NewGridWithColumns(3)
	for _, mod := range mods {
		modCopy := mod
		var img fyne.CanvasObject = widget.NewLabel("Нет изображения")
		if mod.ImageURL != "" {
			if uri, err := url.Parse(mod.ImageURL); err == nil {
				image := canvas.NewImageFromURI(storage.NewURI(uri.String()))
				image.FillMode = canvas.ImageFillContain
				image.SetMinSize(fyne.NewSize(150, 150))
				img = image
			}
		}

		card := container.NewVBox(
			img,
			widget.NewLabel(mod.Name),
			widget.NewButton("Удалить", func() {
				confirm := dialog.NewConfirm("Удалить мод", "Вы уверены?", func(confirmed bool) {
					if !confirmed {
						return
					}
					err := installlog.DeleteInstalledMod(modCopy.ID, dir)
					if err != nil {
						dialog.ShowError(fmt.Errorf("ошибка при удалении: %w", err), window)
						return
					}
					window.Close()
					showInstalledModsWindow(a, parent, dir)
				}, window)
				confirm.Show()
			}),
		)

		grid.Add(container.NewBorder(nil, nil, nil, nil, card))
	}

	scroll := container.NewVScroll(grid)
	window.SetContent(scroll)
	window.Show()
}

func showModsWindow(a fyne.App, parent fyne.Window, initialMods []gamebanana.Mod, saveDir string) {
	modsWindow := a.NewWindow("Доступные моды")
	modsWindow.Resize(fyne.NewSize(800, 600))

	currentPage := 1
	var items []fyne.CanvasObject

	grid := container.NewGridWithColumns(3)
	scroll := container.NewVScroll(grid)

	loadMoreBtn := widget.NewButton("Загрузить ещё", func() {
		progressBar := widget.NewProgressBarInfinite()
		progressBar.Start()
		customDialog := dialog.NewCustomWithoutButtons("Загрузка", progressBar, modsWindow)
		customDialog.Show()

		go func() {
			currentPage++
			newMods, err := gamebanana.FetchMods(currentPage)
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
		*items = append(*items, card)
		grid.Add(container.NewBorder(nil, nil, nil, nil, card))
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

		_ = installlog.SaveInstalledMod(installlog.InstalledMod{
			ID:        mod.ID,
			Name:      mod.Name,
			ImageURL:  mod.ImageURL(),
			Path:      modPath,
			Installed: time.Now(),
		}, dir)

		fyne.Do(func() {
			progress.Hide()
			dialog.ShowInformation("Успех", fmt.Sprintf("Мод %s установлен успешно", mod.Name), parent)
		})
	}()
}
