package installlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type InstalledMod struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ImageURL  string    `json:"image_url"`
	Path      string    `json:"path"` // путь к установленному VPK-файлу (название файла)
	Installed time.Time `json:"installed"`
}

var logFileName = "installed_mods.json"

// SaveInstalledMod сохраняет информацию об установленном моде в файл
func SaveInstalledMod(mod InstalledMod, dir string) error {
	filePath := filepath.Join(dir, logFileName)

	var mods []InstalledMod
	if data, err := os.ReadFile(filePath); err == nil {
		_ = json.Unmarshal(data, &mods)
	}

	mods = append(mods, mod)

	data, err := json.MarshalIndent(mods, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
func DeleteInstalledMod(id int, dir string) error {
	filePath := filepath.Join(dir, "installed_mods.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var mods []InstalledMod
	if err := json.Unmarshal(data, &mods); err != nil {
		return err
	}

	var updated []InstalledMod
	var deletePath string
	for _, m := range mods {
		if m.ID == id {
			deletePath = m.Path
			continue
		}
		updated = append(updated, m)
	}

	// Удаляем файл мода
	if deletePath != "" {
		_ = os.RemoveAll(deletePath)
	}

	newData, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, newData, 0644)
}
func LoadInstalledMods(dir string) ([]InstalledMod, error) {
	filePath := filepath.Join(dir, "installed_mods.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	var mods []InstalledMod
	err = json.Unmarshal(data, &mods)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора JSON: %w", err)
	}
	return mods, nil
}
