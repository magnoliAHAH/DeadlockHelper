package updater

import (
	"errors"
	"os"
	"regexp"
)

func Update(inputDir string) error {
	filePath := inputDir + "/game/citadel/gameinfo.gi"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("файл gameinfo.gi не найден в указанной директории")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	newContent := `
			SearchPaths {
				Mod 		citadel 
				Write		citadel 
				Game 		citadel/addons 
				Game 		citadel 
				Game 		core 
			}`

	re := regexp.MustCompile(`(?s)SearchPaths\s*{.*?}`)
	updatedData := re.ReplaceAllString(string(data), newContent)

	err = os.WriteFile(filePath, []byte(updatedData), 0644)
	if err != nil {
		return err
	}

	return nil // Успешно
}
