package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"

	"github.com/0829why/svrframe/logx"
)

func GetHmacSha1(data, key string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func GetHmacSha256(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))

	return hex.EncodeToString(mac.Sum(nil))
}

func GetMD5(data string) string {
	h := md5.New()
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}
func GCMEncrypt(text, secretKey string) (string, error) {
	key, _ := hex.DecodeString(secretKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		logx.ErrorF("GCMEncrypt NewCipher err = %v", err)
		return "", err
	}
	aeaGcm, err := cipher.NewGCM(block)
	if err != nil {
		logx.ErrorF("GCMEncrypt NewGCM err = %v", err)
		return "", err
	}
	nonce := make([]byte, aeaGcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logx.ErrorF("GCMEncrypt ReadFull err = %v", err)
		return "", err
	}
	cipherText := aeaGcm.Seal(nonce, nonce, []byte(text), nil)
	encoded := base64.StdEncoding.EncodeToString(cipherText)

	return encoded, nil
}
