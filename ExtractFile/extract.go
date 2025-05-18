package extractfile

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractAndInstallVPK(zipPath string, rootPath string) error {
	fmt.Println("Starting extraction for:", zipPath)

	// 1. Создаём временную папку
	tmpDir, err := os.MkdirTemp("", "mod_extract_")
	if err != nil {
		fmt.Println("Failed to create temp dir:", err)
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	fmt.Println("Created temp dir:", tmpDir)
	defer func() {
		fmt.Println("Removing temp dir:", tmpDir)
		os.RemoveAll(tmpDir)
	}()

	// 2. Открываем архив
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		fmt.Println("Failed to open zip:", err)
		return fmt.Errorf("failed to open zip: %w", err)
	}
	fmt.Println("Opened zip archive")

	// Распаковываем все файлы
	for _, f := range r.File {
		path := filepath.Join(tmpDir, f.Name)
		fmt.Println("Processing file in archive:", f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				r.Close()
				fmt.Println("Failed to create directory:", path, err)
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
			fmt.Println("Created directory:", path)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			r.Close()
			fmt.Println("Failed to create directory for file:", path, err)
			return fmt.Errorf("failed to create directory for file %s: %w", path, err)
		}

		rc, err := f.Open()
		if err != nil {
			r.Close()
			fmt.Println("Failed to open file in archive:", f.Name, err)
			return fmt.Errorf("failed to open file %s in archive: %w", f.Name, err)
		}

		dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			r.Close()
			fmt.Println("Failed to create file:", path, err)
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}

		_, err = io.Copy(dstFile, rc)

		dstFile.Close()
		rc.Close()

		if err != nil {
			r.Close()
			fmt.Println("Failed to copy file:", path, err)
			return fmt.Errorf("failed to copy file %s: %w", path, err)
		}
		fmt.Println("Extracted file:", path)
	}

	if err := r.Close(); err != nil {
		fmt.Println("Failed to close zip reader:", err)
		return fmt.Errorf("failed to close zip reader: %w", err)
	}
	fmt.Println("Closed zip archive")

	// 3. Ищем .vpk файл
	var vpkPath string
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error during filepath walk:", err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".vpk") {
			vpkPath = path
			fmt.Println("Found .vpk file:", vpkPath)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking extracted files:", err)
		return fmt.Errorf("error walking extracted files: %w", err)
	}
	if vpkPath == "" {
		fmt.Println("No .vpk file found in archive")
		return errors.New("no .vpk file found in archive")
	}

	// 4. Копируем .vpk
	addonsDir := filepath.Join(rootPath, "game", "citadel", "addons")
	if err := os.MkdirAll(addonsDir, 0755); err != nil {
		return fmt.Errorf("failed to create addons dir: %w", err)
	}

	destPath := filepath.Join(addonsDir, filepath.Base(vpkPath))

	src, err := os.Open(vpkPath)
	if err != nil {
		fmt.Println("Failed to open extracted vpk file:", err)
		return fmt.Errorf("failed to open extracted vpk file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		fmt.Println("Failed to create destination vpk file:", err)
		return fmt.Errorf("failed to create destination vpk file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		fmt.Println("Failed to copy vpk file:", err)
		return fmt.Errorf("failed to copy vpk file: %w", err)
	}
	fmt.Println("Copied .vpk file to:", destPath)

	// 5. Удаляем zip
	if err := os.Remove(zipPath); err != nil {
		fmt.Println("Failed to delete zip:", err)
		return fmt.Errorf("failed to delete zip: %w", err)
	}
	fmt.Println("Deleted zip file:", zipPath)

	return nil
}
