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
	"github.com/panjf2000/ants/v2"
)

/*
暂时只考虑200M以内的文件,超过无法同步
后期再考虑文件断点续传功能
	先同步一个文件信息的缓存文件 如果中断根据这个缓存文件判断是否文件变化 有变化就要重新同步
	缓存文件保存源文件的大小跟md5 以及已经传输到了哪一分块 总共多少个
	续传的时候把分块继续传上来 传完后合并成完整的文件
*/

func SyncFile(srcPath, dstPath string) error {
	log.Debug("开始同步", srcPath)
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	// 优先判断是否存在目标文件信息
	var dstInfo CacheFileInfo
	dstInfoC, ok := CacheFiles.Load(dstPath)
	if !ok {
		log.Debug("不存在缓存文件信息,全部重新获取")
		//全部重新获取 先获取是否存在
		_, err := os.Stat(dstPath)
		if err == nil {
			dstInfo = CacheFileInfo{
				Md5Dst: FileMd5(dstPath),
				Md5Src: FileMd5(srcPath),
			}
			CacheFiles.Store(dstPath, dstInfo)
		}
	} else {
		log.Debug("存在缓存文件信息,直接使用")
		dstInfo = dstInfoC.(CacheFileInfo)
	}
	if FileMd5(srcPath) == dstInfo.Md5Src {
		log.Debug("文件md5一致")
		return nil
	}
	//开始复制
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	log.Debug("开始复制文件", srcPath)
	dstFile, err := os.Create(dstPath)
	if err != nil {
		log.Debug(err)
		return err
	}
	defer dstFile.Close()
	if !CryptOk {
		//不加密的情况下 直接copy
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
		err = os.Chtimes(dstPath, srcInfo.ModTime(), srcInfo.ModTime())
		if err != nil {
			return err
		}
	} else {
		//加密的情况 调用加密函数
		if err := EncryptFile([]byte(Cryptkey), srcFile, dstFile); err != nil {
			return err
		}
	}
	log.Debug("同步完成", srcPath)
	//同步完成后需要更新缓存信息
	dstInfo = CacheFileInfo{
		Md5Dst: FileMd5(dstPath),
		Md5Src: FileMd5(srcPath),
	}
	CacheFiles.Store(dstPath, dstInfo)
	return nil
}

// SyncDir 同步文件夹
func SyncDir(src, dst string) error {
	// 获取源文件夹信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		log.Debug(err, src)
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
	//启用ants
	defer ants.Release()
	SyncFileTask, _ := ants.NewPool(10)
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
			//缓存一份文件夹信息
			if _, ok := CacheDirs.Load(dstPath); !ok {
				_, err := os.Stat(dstPath)
				if err != nil {
					//if os.IsNotExist(err) { //其他错误也重新创建
					log.Debug("开始创建", dstPath)
					if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
						log.Debug(err)
						return err
					}
					//}
				}
				CacheDirs.Store(dstPath, struct{}{})
			}
			//log.Debug("加入监控", path)
			Watcher.Add(path) //加入监控
			return nil
		}
		//log.Debug("如果是文件，复制文件")
		// 如果是文件，复制文件
		if !info.Mode().IsRegular() {
			return nil
		}
		if err := SyncFileTask.Submit(
			func() {
				if err := SyncFile(path, dstPath); err != nil {
					log.Debug(err)
				}
			}); err != nil {
			log.Debug(err)
		}
		return nil
	})
}
