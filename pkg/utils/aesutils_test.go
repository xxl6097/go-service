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
	raw, _ := Aes256GcmEncrypt(jsonData, KEY)
	fmt.Println(string(raw))
	plain, _ := Aes256GcmDecrypt(raw, KEY)
	fmt.Println(string(plain))
}
