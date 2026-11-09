package ukey

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

// withBuffer 临时注册指定长度的占位区，测试结束后还原
func withBuffer(t *testing.T, size int) {
	t.Helper()
	old := GetBuffer()
	Register(bytes.Repeat([]byte{B}, size))
	t.Cleanup(func() { Register(old) })
}

func placeholder(n int) []byte {
	return bytes.Repeat([]byte{B}, n)
}

// fakeBin 构造一个"类可执行文件"：前缀 + 占位区 + 后缀
func fakeBin(prefix, bufSize, suffix int) []byte {
	out := make([]byte, 0, prefix+bufSize+suffix)
	for i := 0; i < prefix; i++ {
		out = append(out, 0xAA)
	}
	out = append(out, placeholder(bufSize)...)
	for i := 0; i < suffix; i++ {
		out = append(out, 0xBB)
	}
	return out
}

// TestGenConfigExactBufferLength 正常配置：输出长度必须与占位区等长
func TestGenConfigExactBufferLength(t *testing.T) {
	for _, size := range []int{1024, 8192} {
		withBuffer(t, size)
		out, err := GenConfig([]byte("hello gfs config"), false)
		if err != nil {
			t.Fatalf("size=%d GenConfig: %v", size, err)
		}
		if len(out) != size {
			t.Fatalf("size=%d GenConfig 输出 %d 字节，期望 %d（必须与占位区等长）", size, len(out), size)
		}
	}
}

// TestGenConfigRejectOverflow 回归：配置超出占位区必须报错，不能静默撑大
// 这是 getHugePageSize panic 的源头（+17 字节导致 ELF 结构错位）
func TestGenConfigRejectOverflow(t *testing.T) {
	withBuffer(t, 1024)
	// 用不可压缩的随机数据，保证 gzip 后仍然超出 1024
	big := make([]byte, 2048)
	if _, err := rand.Read(big); err != nil {
		t.Fatalf("rand: %v", err)
	}
	out, err := GenConfig(big, false)
	if err == nil {
		t.Fatalf("配置溢出时必须报错，实际返回 %d 字节（占位区 1024）", len(out))
	}
	if len(out) != 0 {
		t.Fatalf("报错时不应返回数据，实际 %d 字节", len(out))
	}
}

// TestGenConfigRejectUnregistered 未注册占位区必须报错
func TestGenConfigRejectUnregistered(t *testing.T) {
	withBuffer(t, 1024)
	Register(nil)
	defer func() { Register(placeholder(1024)) }()
	if _, err := GenConfig([]byte("x"), false); err == nil {
		t.Fatal("占位区未注册时应报错")
	}
}

// TestGenerateBinRejectLengthMismatch 回归：长度不等必须拒绝替换
func TestGenerateBinRejectLengthMismatch(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	content := fakeBin(128, 1024, 128)
	if err := os.WriteFile(src, content, 0755); err != nil {
		t.Fatalf("write src: %v", err)
	}
	dst := filepath.Join(dir, "dst.bin")
	// 复现历史场景：占位区 1024，写入 1041（多 17 字节）
	newBytes := make([]byte, 1041)
	for i := range newBytes {
		newBytes[i] = 0x11
	}
	if err := GenerateBin(src, dst, placeholder(1024), newBytes); err == nil {
		t.Fatal("长度不一致时必须报错")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("失败时不应留下产物: %s", dst)
	}
	after, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read src: %v", err)
	}
	if !bytes.Equal(after, content) {
		t.Fatal("源文件不应被修改")
	}
}

// TestSignLoadRoundTrip 全链路：GenConfig -> GenerateBin -> GetCfgBufferFromFile -> LoadBuffer
func TestSignLoadRoundTrip(t *testing.T) {
	const size = 8192
	withBuffer(t, size)

	want := bytes.Repeat([]byte("配置文件内容-"), 20)
	signed, err := GenConfig(want, false)
	if err != nil {
		t.Fatalf("GenConfig: %v", err)
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	content := fakeBin(256, size, 256)
	if err := os.WriteFile(src, content, 0755); err != nil {
		t.Fatalf("write src: %v", err)
	}
	dst := filepath.Join(dir, "dst.bin")
	if err := GenerateBin(src, dst, placeholder(size), signed); err != nil {
		t.Fatalf("GenerateBin: %v", err)
	}

	// 1) 文件长度必须不变
	st, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if st.Size() != int64(len(content)) {
		t.Fatalf("签名后长度 %d，期望 %d（长度变化会破坏可执行文件）", st.Size(), len(content))
	}

	// 2) 从产物中提取签名数据
	extracted := GetCfgBufferFromFile(dst)
	if extracted == nil {
		t.Fatal("GetCfgBufferFromFile 返回 nil")
	}

	// 3) 解析回配置
	body, err := LoadBuffer(extracted)
	if err != nil {
		t.Fatalf("LoadBuffer: %v", err)
	}
	if !bytes.Equal(body, want) {
		t.Fatalf("配置不一致: %q", string(body))
	}
}

// TestUpgradeAcrossBufferSize 跨占位区长度升级：
// 旧二进制占位区 1024 的签名数据，必须能补齐后写进占位区 8192 的新二进制
func TestUpgradeAcrossBufferSize(t *testing.T) {
	dir := t.TempDir()
	want := []byte("legacy config from v1")

	// --- 旧版本：占位区 1024 ---
	withBuffer(t, 1024)
	oldSigned, err := GenConfig(want, false)
	if err != nil {
		t.Fatalf("旧版本 GenConfig: %v", err)
	}
	oldBin := filepath.Join(dir, "old.bin")
	if err := os.WriteFile(oldBin, fakeBin(64, 1024, 64), 0755); err != nil {
		t.Fatalf("write old: %v", err)
	}
	oldSignedBin := filepath.Join(dir, "old_signed.bin")
	if err := GenerateBin(oldBin, oldSignedBin, placeholder(1024), oldSigned); err != nil {
		t.Fatalf("旧版本 GenerateBin: %v", err)
	}

	// 旧文件里提取出的签名数据（长度由文件内长度头决定，与当前 buffer 无关）
	extracted := GetCfgBufferFromFile(oldSignedBin)
	if extracted == nil {
		t.Fatal("旧文件提取签名数据失败")
	}
	// 提取结果 = rawKey + 4 + body，不含尾部 '+' 补齐区
	if len(extracted) > len(oldSigned) {
		t.Fatalf("提取长度 %d 超出旧签名数据 %d", len(extracted), len(oldSigned))
	}
	if !bytes.Equal(extracted, oldSigned[:len(extracted)]) {
		t.Fatal("提取的签名数据与原始签名数据不一致")
	}

	// --- 新版本：占位区 8192 ---
	withBuffer(t, 8192)
	normalized, err := NormalizeBuffer(extracted)
	if err != nil {
		t.Fatalf("NormalizeBuffer: %v", err)
	}
	if len(normalized) != 8192 {
		t.Fatalf("补齐后长度 %d，期望 8192", len(normalized))
	}

	newBin := filepath.Join(dir, "new.bin")
	if err := os.WriteFile(newBin, fakeBin(64, 8192, 64), 0755); err != nil {
		t.Fatalf("write new: %v", err)
	}
	newSignedBin := filepath.Join(dir, "new_signed.bin")
	if err := GenerateBin(newBin, newSignedBin, placeholder(8192), normalized); err != nil {
		t.Fatalf("新版本 GenerateBin: %v", err)
	}

	// 升级后配置必须完好
	data, err := os.ReadFile(newSignedBin)
	if err != nil {
		t.Fatalf("read new: %v", err)
	}
	if len(data) != 64+8192+64 {
		t.Fatalf("签名后长度 %d，期望 %d", len(data), 64+8192+64)
	}
	body, err := LoadBuffer(GetCfgBufferFromFile(newSignedBin))
	if err != nil {
		t.Fatalf("升级后 LoadBuffer: %v", err)
	}
	if !bytes.Equal(body, want) {
		t.Fatalf("升级后配置不一致: %q", string(body))
	}
}

// TestNormalizeBufferTooLong 超长数据必须报错
func TestNormalizeBufferTooLong(t *testing.T) {
	withBuffer(t, 1024)
	if _, err := NormalizeBuffer(make([]byte, 2048)); err == nil {
		t.Fatal("超长数据应报错")
	}
}
