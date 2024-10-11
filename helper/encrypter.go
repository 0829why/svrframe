package helper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

// ECBEncrypter 实现 AES-256-ECB 加密模式
type ECBEncrypter struct {
	b         cipher.Block
	blockSize int
}

// NewECBEncrypter 创建一个 AES-256-ECB 加密模式的 ECBEncrypter
func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return &ECBEncrypter{b, b.BlockSize()}
}

// BlockSize 返回块的大小
func (x *ECBEncrypter) BlockSize() int { return x.blockSize }

// CryptBlocks 加密块
func (x *ECBEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

func EncryptAES(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ecb := NewECBEncrypter(block)
	paddedPlaintext := pad(plaintext, block.BlockSize())
	encrypted := make([]byte, len(paddedPlaintext))
	ecb.CryptBlocks(encrypted, []byte(paddedPlaintext))

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// ECBDecrypter 实现 AES-256-ECB 解密模式
type ECBDecrypter struct {
	b         cipher.Block
	blockSize int
}

// NewECBDecrypter 创建一个 AES-256-ECB 解密模式的 ECBDecrypter
func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return &ECBDecrypter{b, b.BlockSize()}
}

// BlockSize 返回块的大小
func (x *ECBDecrypter) BlockSize() int { return x.blockSize }

// CryptBlocks 解密块
func (x *ECBDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

// pkcs7 pad填充
func pad(s string, blockSize int) []byte {
	padding := blockSize - len(s)%blockSize
	padText := []byte{byte(padding)}
	padText = append(padText, bytes.Repeat([]byte{byte(padding)}, padding-1)...)
	return append([]byte(s), padText...)
}

// pkcs7 pad解除
func unpad(s []byte) []byte {
	length := len(s)
	unpadding := int(s[length-1])
	return s[:length-unpadding]
}

func DecryptAES(key []byte, ciphertext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(decodedCiphertext))
	ecb := NewECBDecrypter(block)
	ecb.CryptBlocks(decrypted, decodedCiphertext)

	return string(unpad(decrypted)), nil
}
