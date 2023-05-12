package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Unzip(zipFile string, targetDir string) error {
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("Unzip(%s): %s", zipFile, err.Error())
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Clean(filepath.Join(targetDir, f.Name))
		if f.FileInfo().IsDir() {
			fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("Unzip(%s): %s", zipFile, err.Error())
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return fmt.Errorf("Unzip(%s): %s", zipFile, err.Error())
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return fmt.Errorf("Unzip(%s): %s", zipFile, err.Error())
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	return nil
}
