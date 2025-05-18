// File: Parser/gamebanana.go
package gamebanana

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// --- структура для списка модов
type Mod struct {
	ID   int    `json:"_idRow"`
	Name string `json:"_sName"`
}

type ApiResponse struct {
	ARecords []Mod `json:"_aRecords"`
}

// FetchMods возвращает первые 10 модов для игры с ID=20948
func FetchMods() ([]Mod, error) {
	const urlMods = "https://gamebanana.com/apiv11/Mod/Index?_nPerpage=10&_nPage=1&_aFilters[Generic_Game]=20948"
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
		FileName    string `json:"_sFile"` // Добавляем!
	} `json:"_aFiles"`
}

// GetDownloadURL получает прямую ссылку на скачивание
func GetDownloadURL(modID int) (string, error) {
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
	if len(data.ARecords) == 0 || data.ARecords[0].DownloadURL == "" {
		return "", errors.New("download URL not found")
	}
	return data.ARecords[0].DownloadURL, nil
}

// DownloadModToDir скачивает файл мода и сохраняет его в папку dir
func DownloadModToDir(modID int, dir string) (string, error) {
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
	fileName := fileInfo.FileName
	if downloadURL == "" || fileName == "" {
		return "", errors.New("incomplete file data")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	outPath := filepath.Join(dir, fileName)
	downloadResp, err := http.Get(downloadURL)
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
