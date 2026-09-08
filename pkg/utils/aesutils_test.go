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
	tesetString := "xiaxiaoli"
	data := TestNode{
		Label:    tesetString,
		Value:    tesetString,
		FilePath: "/Users/xiaxiaoli/Desktop/test.txt",
	}
	jsonData, _ := json.Marshal(data)
	fmt.Println(string(jsonData))
	plarin := Set(string(jsonData))
	fmt.Println(plarin)

	fmt.Println(Get(plarin))
}
