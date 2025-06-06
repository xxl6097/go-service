package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/xxl6097/go-service/pkg/utils"
	"github.com/xxl6097/go-update/pkg/binarydist"
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

	oldFiles, err := os.ReadDir(oldDir)
	if err == nil {
		fmt.Println("旧版本目录：")
		for _, oldFile := range oldFiles {
			fmt.Println(oldFile.Name())
		}
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
		s, e := diff(oldFilePath, filepath.Join(newDir, newFileName))
		if e != nil {
			fmt.Printf("生产差分包失败 %s %v\n", newFileName, e)
		} else {
			fmt.Printf("生产差分包成功 %s\n", s)
		}
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
		newFileName := chgOldFileName(oldFileName, version)
		err = os.Rename(filepath.Join(oldDir, oldFileName), filepath.Join(oldDir, newFileName))
		if err != nil {
			fmt.Printf("修改名称失败 %s-->%s\n", oldFileName, newFileName)
		} else {
			fmt.Printf("修改名称成功 %s-->%s\n", oldFileName, newFileName)
		}

	}
	return nil
}

func chgOldFileName(filename, v string) string {
	re := regexp.MustCompile(`_v\d+\.\d+\.\d+_`)
	newName := re.ReplaceAllString(filename, fmt.Sprintf("_%s_", v)) // 替换为单个下划线
	return newName
}
func diff(older, newer string) (string, error) {
	oldFile, err := os.Open(older)
	if err != nil {
		return "", err
	}
	newFile, err := os.Open(newer)
	if err != nil {
		return "", err
	}
	patch := new(bytes.Buffer)
	err = binarydist.Diff(bufio.NewReader(oldFile), bufio.NewReader(newFile), patch)
	if err != nil {
		return "", err
	}
	pathName := fmt.Sprintf("%s.patch", filepath.Base(newer))
	patchPath := filepath.Join(filepath.Dir(newer), pathName)
	return patchPath, os.WriteFile(patchPath, patch.Bytes(), 0644)
}
