package gamebanana

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// --- структура для списка модов
type Mod struct {
	ID    int    `json:"_idRow"`
	Name  string `json:"_sName"`
	Media Media  `json:"_aPreviewMedia"`
}

type ApiResponse struct {
	ARecords []Mod `json:"_aRecords"`
}

type Media struct {
	Images []Image `json:"_aImages"`
}

type Image struct {
	BaseURL string `json:"_sBaseUrl"`
	File220 string `json:"_sFile220"`
}

func (m Mod) ImageURL() string {
	if len(m.Media.Images) > 0 {
		img := m.Media.Images[0]
		if img.BaseURL != "" && img.File220 != "" {
			return fmt.Sprintf("%s/%s", img.BaseURL, img.File220)
		}
	}
	return ""
}

// FetchMods возвращает первые 20 модов для игры с ID=20948
func FetchMods(page int) ([]Mod, error) {
	urlMods := fmt.Sprintf("https://gamebanana.com/apiv11/Mod/Index?_nPerpage=20&_nPage=%d&_aFilters[Generic_Game]=20948", page)
	resp, err := http.Get(urlMods)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.ARecords, nil
}

// --- структура для получения ссылки на файл
type ModFilesResponse struct {
	ARecords []struct {
		DownloadURL string `json:"_sDownloadUrl"`
		FileName    string `json:"_sFile"` // e.g. "pak25_dir.vpk"
	} `json:"_aFiles"`
}

// DownloadModToDir скачивает VPK файл мода и сохраняет его в папку dir.
// Если внутри директории уже есть файлы вида prefixNN[_suffix].vpk, то новый будет назван с номером на 1 больше.
func DownloadModToDir(modID int, dir string) (string, error) {
	// Запрос к API за данными файла
	apiURL := fmt.Sprintf("https://gamebanana.com/apiv11/Mod/%d?_csvProperties=_aFiles", modID)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected API status: %s", resp.Status)
	}

	var data ModFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("JSON decode error: %w", err)
	}
	if len(data.ARecords) == 0 {
		return "", errors.New("no files found for mod")
	}

	fileInfo := data.ARecords[0]
	downloadURL := fileInfo.DownloadURL
	fileName := fileInfo.FileName // e.g. "pak25_dir.vpk"
	if downloadURL == "" || fileName == "" {
		return "", errors.New("incomplete file data")
	}

	// Создаём директорию, если нет
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// Регулярка с учётом необязательного суффикса, до 99
	// Группы: 1-prefix, 2-num, 3-suffix (например "_dir"), 4-ext
	re := regexp.MustCompile(`^([a-zA-Z]+)(\d{1,2})(_[^\.]+)?(\.[^.]+)$`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) != 5 {
		// Не удалось распарсить — сохраняем оригинал
		return downloadAndSave(downloadURL, filepath.Join(dir, fileName))
	}
	prefix := matches[1]
	suffix := matches[3] // может быть "" или "_dir"
	ext := matches[4]

	// Собираем существующие номера
	nums := []int{}
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		sub := re.FindStringSubmatch(f.Name())
		if len(sub) == 5 && sub[1] == prefix && sub[3] == suffix {
			var n int
			fmt.Sscanf(sub[2], "%d", &n)
			nums = append(nums, n)
		}
	}
	// Находим следующий номер
	sort.Ints(nums)
	next := 1
	if len(nums) > 0 {
		next = nums[len(nums)-1] + 1
	}
	if next > 99 {
		return "", fmt.Errorf("too many files for prefix %s%s", prefix, suffix)
	}
	newName := fmt.Sprintf("%s%02d%s%s", prefix, next, suffix, ext)
	outPath := filepath.Join(dir, newName)

	return downloadAndSave(downloadURL, outPath)
}

// downloadAndSave скачивает по URL и сохраняет в указанный путь
func downloadAndSave(url, outPath string) (string, error) {
	downloadResp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: %s", downloadResp.Status)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, downloadResp.Body); err != nil {
		return "", err
	}

	return outPath, nil
}
