package main

import (
	"litgame.cn/Server/Core/Modules/GameServer"
	"os"
	"strconv"
)

func main() {

	sid := os.Args[1]
	sidint, err := strconv.Atoi(sid)
	if err != nil {
		return
	}

	gameServer := GameServer.CreateGameServer()

	gameServer.Start("", int32(sidint))
}
