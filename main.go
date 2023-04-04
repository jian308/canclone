/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 13:12:07
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-03 22:45:00
 */
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jian308/canclone/clone"
	"github.com/jian308/canclone/webapi"
	"github.com/jian308/go/conf"
	"github.com/jian308/go/log"
	"go.uber.org/zap/zapcore"
)

func main() {
	conf.Auto() //加载配置文件
	if conf.Get("clone.debug") != nil && conf.Get("clone.debug").(bool) {
		log.SetLevel(zapcore.DebugLevel)
	} else {
		log.SetLevel(zapcore.InfoLevel)
	}
	//启动接口服务
	go webapi.WebNew()
	//开启文件同步系统
	go clone.New()
	log.Info("服务启动成功!")
	//优雅关闭开启的服务
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	siger := <-c
	if siger == syscall.SIGINT {
		//方便调试的时候直接关闭
		log.Infof("ctrl+c直接关闭:%d", syscall.Getpid())
		if err := clone.Close(); err != nil {
			log.Error(err)
		}
		os.Exit(0)
	}
	if siger == syscall.SIGTERM {
		log.Infof("开启无忧关闭...等待处理完请求将关闭进程id:%d", syscall.Getpid())
		//app.Shutdown()会等待一个新请求处理完后关闭
		if err := webapi.WebStop(); err != nil {
			log.Error(err)
		}
		if err := clone.Close(); err != nil {
			log.Error(err)
		}
		log.Info("成功关闭!")
	}
}
