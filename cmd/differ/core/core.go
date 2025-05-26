package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/kr/binarydist"
	"github.com/xxl6097/go-service/pkg/utils"
	"os"
	"path/filepath"
	"regexp"
)

func Diff(oldDir string, newDir, version string) error {
	if oldDir == "" || newDir == "" || version == "" {
		return errors.New("oldDir or newDir is empty")
	}
	err := chgOlderToNewer(oldDir, version)
	if err != nil {
		return err
	}
	newFiles, err := os.ReadDir(newDir)
	if err != nil {
		return err
	}
	for _, newFile := range newFiles {
		if newFile.IsDir() {
			continue
		}
		newFileName := newFile.Name()
		oldFilePath := filepath.Join(oldDir, newFileName)
		if !utils.FileExists(oldFilePath) {
			continue
		}
		_ = diff(oldFilePath, filepath.Join(newDir, newFileName))
	}
	return err
}

func chgOlderToNewer(oldDir, version string) error {
	if oldDir == "" || version == "" {
		return errors.New("oldDir or newDir is empty")
	}
	oldFiles, err := os.ReadDir(oldDir)
	if err != nil {
		return err
	}
	for _, oldFile := range oldFiles {
		if oldFile.IsDir() {
			continue
		}
		oldFileName := oldFile.Name()
		_ = os.Rename(filepath.Join(oldDir, oldFileName), filepath.Join(oldDir, chgOldFileName(oldFileName, version)))
	}
	return nil
}

func chgOldFileName(filename, v string) string {
	re := regexp.MustCompile(`_v\d+\.\d+\.\d+_`)
	newName := re.ReplaceAllString(filename, fmt.Sprintf("_%s_", v)) // 替换为单个下划线
	fmt.Println(newName)
	return newName
}
func diff(older, newer string) error {
	oldFile, err := os.Open(older)
	if err != nil {
		return err
	}
	newFile, err := os.Open(newer)
	if err != nil {
		return err
	}
	patch := new(bytes.Buffer)
	err = binarydist.Diff(bufio.NewReader(oldFile), bufio.NewReader(newFile), patch)
	if err != nil {
		return err
	}
	fileName := filepath.Base(newer)
	return os.WriteFile(filepath.Join(filepath.Dir(newer), fileName), patch.Bytes(), 0644)
}
