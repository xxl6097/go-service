// Package aesutil 提供仅 AES-256-GCM 的跨平台加解密工具函数。
//
// 纯 Go 标准库实现，零第三方依赖、无 cgo，可直接编译运行于
// macOS / Linux / Windows，且密文格式与平台无关：
//
//	密文 = nonce(12 字节) || AES-256-GCM 密文(含 16 字节认证标签)
//
// 用法（形如静态工具类，直接调用包级函数）:
//
//	key, _ := aesutil.NewKey()                    // 生成 32 字节随机密钥
//	ct, err := aesutil.Encrypt([]byte("机密"), key) // 加密：AES-256-GCM
//	pt, err := aesutil.Decrypt(ct, key)            // 解密
//
// 说明:
//   - key 必须恰好 32 字节（AES-256），否则返回 ErrBadKeyLength
//   - 每次 Encrypt 自动生成随机 nonce，相同明文每次密文不同
//   - GCM 自带认证：密钥错误或密文被篡改时返回 ErrDecrypt，可 errors.Is 判定
//   - 密钥可用 KeyFromHex 由 64 位 hex 解析（便于配置文件保存）
package aesutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

const (
	// KeySize AES-256 密钥字节数。
	KeySize = 32
	// NonceSize GCM nonce 推荐长度。
	NonceSize = 12
	// TagSize GCM 认证标签字节数。
	TagSize = 16
	// Overhead 密文相对明文的固定增量：nonce + 认证标签。
	Overhead = NonceSize + TagSize
)

var (
	// ErrBadKeyLength 密钥长度不是 32 字节。
	ErrBadKeyLength = errors.New("AES-256 密钥必须为 32 字节")
	// ErrDecrypt 解密失败：密钥错误或数据被篡改（GCM 认证失败）。
	ErrDecrypt = errors.New("密钥错误或数据被篡改")
)

// Encrypt 用 AES-256-GCM 加密明文。
// 返回 data = nonce(12B) || ciphertext(含 16B 认证标签)。
func Encrypt(plain, key []byte) ([]byte, error) {
	gcm, err := gcmCipher(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	out := make([]byte, 0, Overhead+len(plain))
	out = append(out, nonce...)
	return gcm.Seal(out, nonce, plain, nil), nil
}

// Decrypt 用 AES-256-GCM 解密 Encrypt 产生的密文。
func Decrypt(data, key []byte) ([]byte, error) {
	gcm, err := gcmCipher(key) // 先校验密钥，再校验数据
	if err != nil {
		return nil, err
	}
	if len(data) < Overhead {
		return nil, ErrDecrypt
	}
	plain, err := gcm.Open(nil, data[:NonceSize], data[NonceSize:], nil)
	if err != nil {
		return nil, ErrDecrypt // GCM 认证失败 = 密钥错误或数据被篡改
	}
	return plain, nil
}

// NewKey 生成 32 字节随机密钥（AES-256）。可配合 KeyHex 转 hex 保存。
func NewKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// KeyFromHex 由 64 位 hex（可带 0x 前缀）解析出 32 字节 AES-256 密钥。
func KeyFromHex(s string) ([]byte, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "0x")
	if len(s) != 2*KeySize {
		return nil, ErrBadKeyLength
	}
	key, err := hex.DecodeString(s)
	if err != nil {
		return nil, ErrBadKeyLength
	}
	return key, nil
}

// KeyHex 将密钥转为 64 位小写 hex 字符串。
func KeyHex(key []byte) string {
	return hex.EncodeToString(key)
}

// gcmCipher 校验密钥并构造 GCM AEAD。
func gcmCipher(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, ErrBadKeyLength
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
