package utils

import (
	"encoding/json"
	"fmt"
	"testing"
)

type TestNode struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	FilePath string `json:"filePath"`
}

func TestEnDeCode(t *testing.T) {
	tesetString := "aaaaa"
	data := TestNode{
		Label:    tesetString,
		Value:    tesetString,
		FilePath: "/Users/aaaa/Desktop/test.txt",
	}
	jsonData, _ := json.Marshal(data)
	fmt.Println(string(jsonData))
	plarin, _ := Set(string(jsonData))
	fmt.Println(plarin)

	fmt.Println(Get(plarin))
}
