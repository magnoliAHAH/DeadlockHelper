package installlog

import (
	"encoding/json"
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
