package updater

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func Update(inputDir string) {
	reader := bufio.NewReader(os.Stdin)

	defer func() {
		fmt.Println("Нажмите Enter для выхода...")
		reader.ReadString('\n')
	}()

	filePath := inputDir + "/game/citadel/gameinfo.gi"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("Файл gameinfo.gi не найден в указанной директории.")

	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Ошибка чтения: ", err)

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
		fmt.Println("Ошибка при записи в файл:", err)
		return
	}

	fmt.Println("Файл успешно обновлён.")

}
