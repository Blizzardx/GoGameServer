package main

import (
	"litgame.cn/Server/Core/Modules/ComponentServer"
	"os"
	"strconv"
)

func main() {
	sid := os.Args[1]
	sidint, err := strconv.Atoi(sid)
	if err != nil {
		return
	}

	componentServer := ComponentServer.CreateComponentServer()

	componentServer.Start("", int32(sidint), 0)
}
