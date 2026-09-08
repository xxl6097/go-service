package ukey

import (
	"fmt"
	"testing"
)

type TestNode struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	FilePath string `json:"filePath"`
}

func TestAesGCM(t *testing.T) {
	tesetString := "aaaa"
	data := TestNode{
		Label:    tesetString,
		Value:    tesetString,
		FilePath: "/Users/aaaaa/Desktop/test.txt",
	}
	data1, _ := StructToAesGcm(&data)
	fmt.Println(string(data1))
}
