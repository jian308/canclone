/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-01 14:06:14
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-01 14:06:18
 */
package clone

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func FileMd5(Path string) string {
	// 计算文件的md5值
	f1, err := os.Open(Path)
	if err != nil {
		return ""
	}
	defer f1.Close()
	h1 := md5.New()
	if _, err := io.Copy(h1, f1); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h1.Sum(nil))
}
