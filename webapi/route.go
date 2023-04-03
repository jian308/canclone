/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 13:35:39
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-03 21:50:21
 */
package webapi

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jian308/canclone/clone"
	"github.com/jian308/go/conf"
	"github.com/jian308/go/log"
	"go.uber.org/zap/zapcore"
)

func Route(app *fiber.App) {
	api := app.Group("/Api", cors.New()) //ApiCheck
	api.Options("/:act", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})
	//接口路由绑定
	//拉取功能,一般用本地没有得时候，用空文件夹从云端恢复文件用
	api.Get("/pull", func(c *fiber.Ctx) error {
		if err := clone.BakDir(clone.DirDst, clone.DirSrc); err != nil {
			return c.SendString(err.Error())
		}
		return c.SendString("拉取完成!")
	})
	api.Get("/lconf", func(c *fiber.Ctx) error {
		//重新读取conf.toml
		conf.Auto()
		//log重载
		if conf.Get("clone.debug") != nil && conf.Get("clone.debug").(bool) {
			log.SetLevel(zapcore.DebugLevel)
		} else {
			log.SetLevel(zapcore.InfoLevel)
		}
		//ticker变动后期支持修改
		return c.SendString("重载conf配置!")
	})
	api.Get("/lfilter", func(c *fiber.Ctx) error {
		clone.LoadFilter()
		return c.SendString("重载filter配置!")
	})
}
