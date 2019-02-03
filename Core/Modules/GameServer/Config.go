package GameServer

import "github.com/Blizzardx/GoGameServer/Core/Modules/Config"

func (server *GameServer) initConfig(remoteConfigUrl string) {

	// fetch config
	server.config = Config.FetchRemoteConfig(remoteConfigUrl)

	for _, gameServer := range server.config.GameServerList {
		if gameServer.LogicId == server.logicId {
			server.selfConfig = gameServer
			break
		}
	}
}
