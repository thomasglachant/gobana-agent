package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	b64 "encoding/base64"
	"fmt"
	"strings"
)

func AESEncrypt(src []byte, key string) (encryptedData []byte, encryptErr error) {
	iv := make([]byte, aes.BlockSize)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(PKCS5Padding([]byte(key), aes.BlockSize))
	if err != nil {
		return nil, err
	}
	if len(src) == 0 {
		return nil, fmt.Errorf("plain content empty")
	}

	ecb := cipher.NewCBCEncrypter(block, iv)
	content := src
	content = PKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)

	return []byte(fmt.Sprintf("%s.%s", b64.StdEncoding.EncodeToString(iv), b64.StdEncoding.EncodeToString(crypted))), nil
}

func AESDecrypt(crypt []byte, key string) (decryptedData []byte, decryptErr error) {
	bodySplitted := strings.Split(string(crypt), ".")
	if len(bodySplitted) != 2 {
		return nil, fmt.Errorf("invalid body")
	}
	iv := bodySplitted[0]
	encryptedData := bodySplitted[1]

	var cryptb64, ivb64 []byte
	var err error
	cryptb64, err = b64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	ivb64, err = b64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(PKCS5Padding([]byte(key), aes.BlockSize))
	if err != nil {
		return nil, err
	}
	if len(crypt) == 0 {
		return nil, fmt.Errorf("cipher content empty")
	}
	ecb := cipher.NewCBCDecrypter(block, ivb64)
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, cryptb64)

	return PKCS5Trimming(decrypted), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
