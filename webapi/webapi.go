/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 13:28:16
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-01 13:39:09
 */
package webapi

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jian308/go/conf"
	"github.com/jian308/go/log"
)

var app *fiber.App

func WebNew() {
	app = fiber.New(fiber.Config{
		DisableStartupMessage: true, //隐藏启动信息
	})
	//路由绑定
	Route(app)
	// 读取端口
	var listen string
	if conf.Get("webapi.listen") != nil {
		listen = conf.Get("webapi.listen").(string)
	} else {
		log.Error("未找到监听端口,启动失败!")
	}
	app.Listen(listen)
}

func WebStop() error {
	return app.Shutdown()
}
