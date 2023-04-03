/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 13:40:26
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-03 21:16:25
 */
package clone

import (
	"path/filepath"

	"github.com/jian308/go/conf"
	"github.com/jian308/go/log"
)

var (
	DirSrc string
	DirDst string
)

func New() {
	LoadFilter()  //初始化加载过滤器
	WatcherInit() //初始化文件监控器
	CryptInit()   //初始化加密器
	CacheLoad()   //加载缓存目录+文件
	// 源文件夹
	src, err := filepath.Abs(conf.Get("clone.dir_src").(string))
	if err != nil {
		log.Fatal("获取源文件夹失败", err)
	}
	DirSrc = src
	log.Debug("源文件夹:", DirSrc)
	// 目标文件夹
	dst, err := filepath.Abs(conf.Get("clone.dir_dst").(string))
	if err != nil {
		log.Fatal("获取目标文件夹失败", err)
	}
	DirDst = dst
	log.Debug("目标文件夹:", DirDst)
	// 同步文件夹
	if err := SyncDir(DirSrc, DirDst); err != nil {
		log.Fatal(err)
	}
	go WatcherGo(DirSrc, DirDst)
	go CacheUpdate()
}

func Close() error {
	//TODO
	CacheSave() //关闭之前保存缓存到文件
	return WatcherClose()
}
