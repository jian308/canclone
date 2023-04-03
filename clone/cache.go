/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 14:13:20
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-03 21:16:19
 */
package clone

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/jian308/go/conf"
	"github.com/jian308/go/log"
)

// 缓存目标文件夹的相关信息
// 本地源文件夹不需要缓存

// 缓存目标文件夹的目录是否存在 直接sync.map 无需极限性能
var CacheDirs sync.Map //文件夹缓存组

// 需要缓存的文件信息 暂时用md5就够用 后期增加其他的
type CacheFileInfo struct {
	Md5Src string //src的md5
	Md5Dst string //dst的md5
}

// 缓存目标文件的信息组  直接sync.map 无需极限性能
var CacheFiles sync.Map //文件缓存组

// json结构
type cachejson struct {
	Dirs  []string                 `json:"dirs"`
	Files map[string]CacheFileInfo `json:"files"`
}

// 静态文件把缓存加载进来
func CacheLoad() {
	jsonf, err := os.ReadFile("./cache.json")
	if err != nil {
		log.Info(err)
		return
	}
	var cachedata cachejson
	if err := json.Unmarshal(jsonf, &cachedata); err == nil {
		for _, dir := range cachedata.Dirs {
			CacheDirs.Store(dir, struct{}{})
		}
		for file, info := range cachedata.Files {
			CacheFiles.Store(file, info)
		}
		log.Debug("加载缓存cache.json成功")
	}
}

// 定时器更新缓存 访问目标文件夹然后全部查一遍缓存里是否存在 不存在就删除
func CacheUpdate() {
	if conf.Get("clone.ticker") == nil {
		log.Info("未发现clone.ticker配置,关闭缓存同步")
		return
	}
	ticker := conf.Get("clone.ticker").(int64)
	if ticker == 0 {
		log.Info("配置关闭缓存同步")
		return
	}
	cachetk := time.NewTicker(time.Duration(ticker) * time.Second)
	for range cachetk.C {
		change := false
		CacheDirs.Range(func(dir, value any) bool {
			info, err := os.Stat(dir.(string))
			if err != nil || !info.IsDir() { //不存在或者不是文件夹的情况就删除掉
				CacheDirs.Delete(dir.(string))
				change = true
			}
			return true
		})
		CacheFiles.Range(func(key, val any) bool {
			notfound := false
			file, ok := key.(string)
			if !ok {
				notfound = true
			}
			info, ok := val.(CacheFileInfo)
			if !ok {
				notfound = true
			}
			if notfound || FileMd5(file) != info.Md5Dst {
				CacheFiles.Delete(file)
				change = true
			}
			return true
		})
		//如果有变化
		if change {
			log.Info("发现目标文件夹有变动,自动同步!")
			// 同步文件夹
			if err := SyncDir(DirSrc, DirDst); err != nil {
				log.Error(err)
			}
		}
		CacheSave() //保存最新缓存
	}
	cachetk.Stop() //暂停
}

// 保存缓存到静态文件
func CacheSave() {
	cachedata := cachejson{
		Dirs:  make([]string, 0),
		Files: make(map[string]CacheFileInfo),
	}
	CacheDirs.Range(func(dir, value any) bool {
		cachedata.Dirs = append(cachedata.Dirs, dir.(string))
		return true
	})
	CacheFiles.Range(func(file, info any) bool {
		cachedata.Files[file.(string)] = info.(CacheFileInfo)
		return true
	})
	jsonf, err := os.Create("./cache.json")
	if err != nil {
		log.Error(err)
		return
	}
	encoder := json.NewEncoder(jsonf)
	if err := encoder.Encode(cachedata); err != nil {
		log.Error("保存缓存cache.json失败", err)
		return
	}
	log.Debug("保存缓存cache.json成功")
}
