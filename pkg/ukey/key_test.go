package ukey

import (
	"reflect"
	"testing"
)

type TestNode struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	FilePath string `json:"filePath"`
}

// TestStructToGobRoundTrip gob 编解码必须可逆
func TestStructToGobRoundTrip(t *testing.T) {
	src := TestNode{
		Label:    "aaaa",
		Value:    "bbbb",
		FilePath: "/tmp/test.txt",
	}
	data, err := StructToGob(src)
	if err != nil {
		t.Fatalf("StructToGob: %v", err)
	}
	var dst TestNode
	if err := GobToStruct(data, &dst); err != nil {
		t.Fatalf("GobToStruct: %v", err)
	}
	if !reflect.DeepEqual(src, dst) {
		t.Fatalf("gob 编解码不一致: %+v != %+v", src, dst)
	}
}

// TestDecKeyRawKey 内置 key 必须能解出原始 rawKey
func TestDecKeyRawKey(t *testing.T) {
	rawKey, err := GetRawKey()
	if err != nil {
		t.Fatalf("GetRawKey: %v", err)
	}
	if len(rawKey) == 0 {
		t.Fatal("rawKey 为空")
	}
}

// TestDecKeyTooShort 长度不足时必须返回错误而不是 panic
func TestDecKeyTooShort(t *testing.T) {
	if _, err := DecKey([]byte{0x01, 0x02, 0x03}); err == nil {
		t.Fatal("key 长度不足时应返回错误")
	}
}
