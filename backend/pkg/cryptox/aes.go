// Package cryptox 提供 AES-256-GCM 加解密工具（ADR-011）。
// 用于身份证号、手机号等敏感字段的加密存储与脱敏展示。
package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// 密钥来源：环境变量 CRYPTOX_KEY，缺省 32 字节开发密钥。
func getKey() []byte {
	if k := os.Getenv("CRYPTOX_KEY"); k != "" {
		if len(k) == 32 {
			return []byte(k)
		}
	}
	// 开发环境默认密钥（32 字节），生产环境必须通过环境变量覆盖
	return []byte("studenthub-dev-aes-key-change-m!")
}

// Encrypt 使用 AES-256-GCM 加密明文，返回 base64(iv|cipher|tag)。
func Encrypt(plain string) string {
	if plain == "" {
		return ""
	}

	key := getKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return ""
	}

	// nonce 作为附加数据前缀，GCM 自带 tag
	cipherText := aesGCM.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(cipherText)
}

// Decrypt 解密 base64(iv|cipher|tag) 密文，返回明文。
func Decrypt(enc string) (string, error) {
	if enc == "" {
		return "", nil
	}

	key := getKey()
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文长度不足")
	}

	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plain, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plain), nil
}

// MaskIDCard 身份证脱敏：保留前 3 位和后 4 位，中间用 * 填充。
// 例：110101200001010023 → 110***********0023
func MaskIDCard(idCard string) string {
	if len(idCard) < 7 {
		return strings.Repeat("*", len(idCard))
	}
	prefix := idCard[:3]
	suffix := idCard[len(idCard)-4:]
	middle := strings.Repeat("*", len(idCard)-7)
	return prefix + middle + suffix
}

// MaskPhone 手机号脱敏：保留前 3 位和后 4 位。
// 例：13812345678 → 138****5678
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
