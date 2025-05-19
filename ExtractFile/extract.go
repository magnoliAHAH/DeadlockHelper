package extractfile

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
	"github.com/nwaples/rardecode"
)

// ExtractAndInstallVPK распаковывает ZIP, RAR или 7z, находит .vpk и устанавливает его в папку addons
func ExtractAndInstallVPK(archivePath string, rootPath string) (string, error) {
	fmt.Println("Starting extraction for:", archivePath)

	// 1. Создаём временную папку
	tmpDir, err := os.MkdirTemp("", "mod_extract_")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() {
		fmt.Println("Removing temp dir:", tmpDir)
		os.RemoveAll(tmpDir)
	}()

	// 2. Распаковываем архив в зависимости от расширения
	switch ext := strings.ToLower(filepath.Ext(archivePath)); ext {
	case ".zip":
		fmt.Println("Detected ZIP archive")
		err = extractZIP(archivePath, tmpDir)
	case ".rar":
		fmt.Println("Detected RAR archive")
		err = extractRAR(archivePath, tmpDir)
	case ".7z":
		fmt.Println("Detected 7z archive")
		err = extract7z(archivePath, tmpDir)
	default:
		err = fmt.Errorf("unsupported archive format: %s", ext)
	}
	if err != nil {
		return "", fmt.Errorf("failed to extract archive: %w", err)
	}

	// 3. Ищем .vpk файл
	var vpkPath string
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".vpk") {
			vpkPath = path
			fmt.Println("Found .vpk file:", vpkPath)
			return io.EOF // прерываем Walk
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error walking extracted files: %w", err)
	}
	if vpkPath == "" {
		return "", errors.New("no .vpk file found in archive")
	}

	// 4. Копируем .vpk
	addonsDir := filepath.Join(rootPath, "game", "citadel", "addons")
	if err := os.MkdirAll(addonsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create addons dir: %w", err)
	}

	destPath := filepath.Join(addonsDir, filepath.Base(vpkPath))
	if err := copyFile(vpkPath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy vpk file: %w", err)
	}
	fmt.Println("Copied .vpk file to:", destPath)

	// 5. Удаляем исходный архив
	if err := os.Remove(archivePath); err != nil {
		return "", fmt.Errorf("failed to delete archive: %w", err)
	}
	fmt.Println("Deleted archive file:", archivePath)

	return destPath, nil
}

// extractZIP распаковывает ZIP архив в указанную папку
func extractZIP(zipPath, dstDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(dstDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", path, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in archive: %w", f.Name, err)
		}

		dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}

		if _, err := io.Copy(dstFile, rc); err != nil {
			rc.Close()
			dstFile.Close()
			return fmt.Errorf("failed to copy file %s: %w", path, err)
		}

		rc.Close()
		dstFile.Close()
	}
	return nil
}

// extractRAR распаковывает RAR архив в указанную папку
func extractRAR(rarPath, dstDir string) error {
	file, err := os.Open(rarPath)
	if err != nil {
		return fmt.Errorf("failed to open rar: %w", err)
	}
	defer file.Close()

	rr, err := rardecode.NewReader(file, "")
	if err != nil {
		return fmt.Errorf("failed to create rar reader: %w", err)
	}

	for {
		hdr, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading rar: %w", err)
		}

		outPath := filepath.Join(dstDir, hdr.Name)
		if hdr.IsDir {
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return fmt.Errorf("failed to create dir in rar: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("failed to create dir for file in rar: %w", err)
		}

		outFile, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create file in rar: %w", err)
		}

		if _, err := io.Copy(outFile, rr); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to extract file from rar: %w", err)
		}

		outFile.Close()
	}
	return nil
}

// extract7z распаковывает 7z архив в указанную папку
func extract7z(archivePath, dstDir string) error {
	r, err := sevenzip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open 7z: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		outPath := filepath.Join(dstDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, f.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", outPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", outPath, err)
		}

		inFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in archive: %w", f.Name, err)
		}

		dstFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			inFile.Close()
			return fmt.Errorf("failed to create file %s: %w", outPath, err)
		}

		if _, err := io.Copy(dstFile, inFile); err != nil {
			inFile.Close()
			dstFile.Close()
			return fmt.Errorf("failed to copy file %s: %w", outPath, err)
		}

		inFile.Close()
		dstFile.Close()
	}
	return nil
}

// copyFile копирует файл из srcPath в dstPath
func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}
	return nil
}
