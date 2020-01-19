package main

/*
 * @Description:
 * @Author: chenwei
 * @Date: 2020-01-19 17:16:43
 */

import (
	"github.com/silenceWe/configgo/server"
)

func main() {
	server.LoadConfig("./cfg_base.ini", ":8080")
}
