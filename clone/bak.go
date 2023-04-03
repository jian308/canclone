/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 14:13:20
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-03 20:03:05
 */
package clone

import (
	"io"
	"os"
	"path/filepath"

	"github.com/jian308/go/log"
)

// 复制SyncDir 同步文件夹
func BakDir(src, dst string) error {
	// 获取源文件夹信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		log.Debug(err)
		return err
	}
	// 创建目标文件夹
	// 判断是否存在
	if _, err := os.Stat(dst); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
				log.Debug(err, dst)
				return err
			}
		}
	}
	// 遍历源文件夹
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debug(err)
			return err
		}
		//过滤
		if DoFilter(path, src) {
			return nil
		}
		// 获取目标文件路径
		dstPath := filepath.Join(dst, path[len(src):])
		// 如果是文件夹，创建目标文件夹
		if info.IsDir() {
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				log.Debug(err)
				return err
			}
			return nil
		}
		//log.Debug("如果是文件，复制文件")
		// 如果是文件，复制文件
		if !info.Mode().IsRegular() {
			return nil
		}
		go func() {
			if err := BakFile(path, dstPath); err != nil {
				log.Debug(err)
			}
		}()
		return nil
	})
}

// 因为是pull拉下来 无需判断是否存在 全部重新拉取
func BakFile(srcPath, dstPath string) error {
	log.Debug("开始bak", srcPath)
	//开始复制
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		log.Debug(err)
		return err
	}
	defer dstFile.Close()
	if !CryptOk {
		log.Debug("未加密,直接copy")
		//不加密的情况下 直接copy
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
	} else {
		log.Debug("已加密,解密copy")
		//加密的情况 调用解密函数
		if err := DecryptFile([]byte(Cryptkey), srcFile, dstFile); err != nil {
			return err
		}
	}
	log.Debug("bak完成", dstPath)
	return nil
}
