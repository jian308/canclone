/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 13:51:36
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-02 21:51:40
 */
package clone

import (
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/jian308/go/log"
)

var Watcher *fsnotify.Watcher

func WatcherInit() {
	// 创建监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return
	}
	Watcher = watcher
}

func WatcherGo(src, dst string) {
	// 添加监听路径
	err := Watcher.Add(src)
	if err != nil {
		log.Fatal(err)
		return
	}
	// 监听文件变化
	for {
		select {
		case event := <-Watcher.Events:
			log.Debug("event", event)
			//计算路径
			srcPath := event.Name
			//过滤不同步
			if DoFilter(srcPath, src) {
				//log.Debug("过滤:", srcPath)
				continue
			}
			dstPath := strings.Replace(srcPath, src, dst, 1)
			//删除文件夹REMOVE|RENAME
			if event.Op&fsnotify.Remove|fsnotify.Rename == fsnotify.Remove|fsnotify.Rename {
				log.Debug("删除文件夹:", srcPath)
				//移除监听目录
				if err := os.RemoveAll(dstPath); err != nil {
					log.Debug(err)
				}
				CacheDirs.Delete(dstPath) //删除文件夹缓存信息
				continue
			}
			//删除文件RENAME
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				log.Debug("删除文件:", srcPath)
				if err := os.Remove(dstPath); err != nil {
					log.Debug(err)
				}
				CacheDirs.Delete(dstPath)
				continue
			}
			//创建文件或文件夹
			if event.Op&fsnotify.Create == fsnotify.Create {
				log.Debug("创建文件:", srcPath)
				info, err := os.Stat(srcPath)
				if err != nil {
					log.Debug(err)
					continue
				}
				//如果是文件夹
				if info.IsDir() {
					//同步创建文件夹
					//缓存一份文件夹信息
					if _, ok := CacheDirs.Load(dstPath); !ok {
						_, err := os.Stat(dstPath)
						if err != nil {
							if os.IsNotExist(err) { //其他错误也重新创建
								log.Debug("开始创建", dstPath)
								if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
									log.Debug(err)
									continue
								}
							} else {
								log.Debug("可能是其他情况")
							}
						}
						CacheDirs.Store(dstPath, struct{}{})
					}
					//后续操作
					//同步创建的文件夹
					if err := SyncDir(srcPath, dstPath); err != nil {
						log.Debug(err)
						continue
					}
					//把同步后的文件夹加入监听文件夹
					err = Watcher.Add(srcPath)
					if err != nil {
						log.Debug(err)
						continue
					}
				} else {
					if err := SyncFile(srcPath, dstPath); err != nil {
						log.Debug(err)
						continue
					}
				}
				log.Debug("创建文件=同步完成")
				continue
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Debug("文件被修改=>", srcPath)
				if err := SyncFile(srcPath, dstPath); err != nil {
					log.Debug(err)
					continue
				}
				log.Debug("文件已同步=>", dstPath)
				continue
			}
		case err := <-Watcher.Errors:
			log.Debug("error:", err)
		}
	}
}

func WatcherClose() error {
	return Watcher.Close()
}
