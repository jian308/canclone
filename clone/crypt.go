/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 14:01:42
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-02 23:19:47
 */
package clone

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"os"

	"github.com/jian308/go/conf"
	"github.com/jian308/go/log"
)

var CryptOk bool
var Cryptkey string

func CryptInit() {
	if conf.Get("clone.passkey") != nil {
		Cryptkey = conf.Get("clone.passkey").(string)
		if len(Cryptkey) < 16 {
			Cryptkey = Cryptkey + "github.com/jian308/canclone"
		}
		if len(Cryptkey) > 16 {
			Cryptkey = Cryptkey[:16]
		}
		CryptOk = true
	}
}

// Define a function to encrypt a file
func EncryptFile(key []byte, inFile, outFile *os.File) error {
	// Generate a new AES cipher block using the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Debug(err)
		return err
	}
	// Generate a new initialization vector
	// 创建 CBC 分组模式实例
	iv := make([]byte, aes.BlockSize)
	stream := cipher.NewOFB(block, iv)
	// 创建加密流
	writer := &cipher.StreamWriter{S: stream, W: outFile}
	// Encrypt the file data and write it to the output file
	if _, err := io.Copy(writer, inFile); err != nil {
		log.Debug(err)
		return err
	}
	return nil
}

// Define a function to decrypt a file
func DecryptFile(key []byte, inFile, outFile *os.File) error {
	// Create a new AES cipher block using the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Debug(err)
		return err
	}
	// 创建 CBC 分组模式实例
	iv := make([]byte, aes.BlockSize)
	stream := cipher.NewOFB(block, iv)
	// 创建解密流
	reader := &cipher.StreamReader{S: stream, R: inFile}
	if _, err := io.Copy(outFile, reader); err != nil {
		log.Debug(err)
		return err
	}
	return nil
}
