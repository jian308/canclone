/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 13:48:28
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-02 02:55:36
 */
package clone

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/jian308/go/log"
)

// 过滤组
var Filters atomic.Value

func LoadFilter() {
	// 读入过滤的文件
	f, err := os.Open("./filter.txt")
	if err != nil {
		log.Info("未发现过滤规则文件filter.txt，关闭过滤器！")
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	filters := make([]string, 0, 1024)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if pattern != "" && !strings.HasPrefix(pattern, "#") {
			filters = append(filters, pattern)
		}
	}
	Filters.Store(filters)
}

func DoFilter(dirName, dir string) bool {
	filters, ok := Filters.Load().([]string)
	if !ok {
		return false
	}
	info, err := os.Stat(dirName)
	if err != nil {
		//log.Debug("获取信息错误", err)
		//一般属于文件不存在的情况 包括重命名/删除等情况 不拦截
		return false
	}
	if info.Size() > 200*1024*1024 {
		log.Debug("拦截超过200M的文件", dirName)
		return true
	}
	for _, v := range filters {
		dirname := dirName
		//fmt.Printf("开始过滤:%s=>%s\n", v, dirname)
		if v[0] == '/' {
			//如果规则开始是/证明要判断根目录固定的,需要完全匹配
			//log.Debug(dirname, filepath.Join(dir, v))
			if dirname == filepath.Join(dir, v) {
				return true
			}
			match, err := filepath.Match(filepath.Join(dir, v), dirname)
			if err != nil {
				log.Debug(err)
				continue
			}
			if match {
				return true
			}
		} else {
			//匹配所有目录里是否存在
			//log.Debugf("所有目录里匹配:v=%s,dirname=%s", v, dirname)
			//是文件且只要找目录的时候
			if !info.IsDir() && v[len(v)-1] == '/' {
				//只保留目录 不用文件名去判断
				dirname = filepath.Dir(dirname)
			}
			dir, file := filepath.Split(dirname)
			dirs := strings.Split(dir, "/")
			dirs = append(dirs, file)
			//留下的路径都是可以找的
			if v[len(v)-1] == '/' {
				v = v[:len(v)-1]
			}
			//log.Debugf("在%+v里搜索%v", dirs, v)
			for _, vd := range dirs {
				//log.Debugf("再匹配单个子目录:v=%s,vd=%s", v, vd)
				match, err := filepath.Match(v, vd)
				if err != nil {
					log.Debug(err)
					continue
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}
