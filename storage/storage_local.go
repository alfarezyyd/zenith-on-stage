package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type LocalStorage struct {
	basePath string
	string
	basePathTrash string
}

func NewLocalStorage(basePath, basePathTrash string) *LocalStorage {
	return &LocalStorage{basePath: basePath, basePathTrash: basePathTrash}
}

func (localStorage *LocalStorage) Put(fileHeader *multipart.FileHeader, path string) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := filepath.Join(localStorage.basePath, path)
	os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return path, err
}

func (localStorage *LocalStorage) MoveFile(originalPath, targetFolder string, isDelete bool) (string, error) {
	srcPath := filepath.Join(localStorage.basePath, originalPath)
	targetDir := filepath.Join(localStorage.basePath, targetFolder)

	if isDelete {
		targetDir = filepath.Join(localStorage.basePathTrash, targetFolder)
	}
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return "", err
	}
	// Ensure target folder exists

	// Create new filename with timestamp to avoid conflicts
	filename := filepath.Base(originalPath)
	newFilename := time.Now().Format("20060102_150405") + "_" + filename
	destPath := filepath.Join(targetDir, newFilename)

	// Move
	if err := os.Rename(srcPath, destPath); err != nil {
		return "", err
	}

	// Return path relative to base
	return filepath.Join(targetFolder, newFilename), nil
}

func (localStorage *LocalStorage) Delete(path string) error {
	return os.Remove(filepath.Join(localStorage.basePath, path))
}

func (localStorage *LocalStorage) Verify(paths []string) ([]string, []string) {
	var invalidPaths []string
	var matchesPaths []string
	for _, path := range paths {
		targetPath := fmt.Sprintf("%s/tmp/%s_*", localStorage.basePath, path)
		matches, _ := filepath.Glob(targetPath)
		if len(matches) == 0 {
			invalidPaths = append(invalidPaths, path)
			continue
		}
		fileName := filepath.Base(matches[0])

		matchesPaths = append(matchesPaths, fileName)
	}
	return matchesPaths, invalidPaths
}
